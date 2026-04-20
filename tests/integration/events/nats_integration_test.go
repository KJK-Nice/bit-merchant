package events_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"bitmerchant/internal/common"
	commonhttp "bitmerchant/internal/common/http"
	"bitmerchant/internal/infrastructure/events"
	"bitmerchant/internal/infrastructure/logging"
	"bitmerchant/internal/infrastructure/repositories/memory"
	orderevent "bitmerchant/internal/ordering/app/event"
	"bitmerchant/internal/ordering/domain/order"
	orderingservice "bitmerchant/internal/ordering/service"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestNATSOrderCreatedBroadcastsToKitchenSSE(t *testing.T) {
	natsURL := setupNATSServer(t)
	logger := logging.NewLogger()
	orderRepo := memory.NewMemoryOrderRepository()
	sseHandler := commonhttp.NewSSEHandler()

	eventBus := newNATSEventBus(t, natsURL, "sse-instance", 1, 800*time.Millisecond)
	t.Cleanup(func() {
		require.NoError(t, eventBus.Close())
	})

	orderID := common.OrderID("order_sse_1")
	orderItem, err := order.NewOrderItem("oi_sse_1", orderID, "item_1", "Burger", 1, 10)
	require.NoError(t, err)
	createdOrder, err := order.NewOrder(
		orderID,
		"1101",
		"restaurant_1",
		"session_1",
		[]order.OrderItem{*orderItem},
		1000,
		common.PaymentMethodTypeCash,
	)
	require.NoError(t, err)
	createdOrder.FiatAmount = 10
	require.NoError(t, orderRepo.Save(createdOrder))

	router := newRouter(t, 5*time.Second)
	router.AddMiddleware(
		middleware.Recoverer,
		middleware.Retry{
			MaxRetries:      3,
			InitialInterval: 25 * time.Millisecond,
			MaxInterval:     250 * time.Millisecond,
			Multiplier:      2,
			Logger:          watermill.NewStdLogger(false, false),
		}.Middleware,
	)
	orderingservice.RegisterOrderSSEHandlers(router, eventBus.Subscriber(), logger, sseHandler, orderRepo)
	runRouter(t, router)

	writer := newSSECaptureWriter()
	reqCtx, cancelReq := context.WithCancel(context.Background())
	defer cancelReq()

	req := httptest.NewRequest(http.MethodGet, "/kitchen/stream", nil).WithContext(reqCtx)
	e := echo.New()
	ctx := e.NewContext(req, writer)

	streamDone := make(chan error, 1)
	go func() {
		streamDone <- sseHandler.KitchenStream(ctx)
	}()

	// Ensure stream subscriber registration before publishing.
	time.Sleep(120 * time.Millisecond)

	require.NoError(t, eventBus.Publish(context.Background(), common.EventOrderCreated, orderevent.OrderCreated{
		OrderID:      createdOrder.ID,
		RestaurantID: createdOrder.RestaurantID,
		OrderNumber:  createdOrder.OrderNumber,
		TotalAmount:  createdOrder.TotalAmount,
		CreatedAt:    createdOrder.CreatedAt,
	}))

	select {
	case chunk := <-writer.writes:
		payload := string(chunk)
		assert.Contains(t, payload, "event: datastar-patch-elements")
		assert.Contains(t, payload, "#orders-list")
	case <-time.After(6 * time.Second):
		t.Fatal("timed out waiting for kitchen SSE message")
	}

	cancelReq()
	select {
	case err := <-streamDone:
		require.NoError(t, err)
	case <-time.After(2 * time.Second):
		t.Fatal("kitchen stream did not stop after cancellation")
	}
}

func TestNATSRedeliveryAfterHandlerError(t *testing.T) {
	natsURL := setupNATSServer(t)

	eventBus := newNATSEventBus(t, natsURL, "retry-instance", 1, 300*time.Millisecond)
	t.Cleanup(func() {
		require.NoError(t, eventBus.Close())
	})

	router := newRouter(t, 5*time.Second)
	var attempts atomic.Int32
	processed := make(chan struct{}, 1)

	router.AddConsumerHandler("retry-once", "test.retry.once", eventBus.Subscriber(), func(msg *message.Message) error {
		attempt := attempts.Add(1)
		if attempt == 1 {
			return errors.New("fail first delivery")
		}
		if attempt == 2 {
			processed <- struct{}{}
		}
		return nil
	})
	runRouter(t, router)

	require.NoError(t, eventBus.Publish(context.Background(), "test.retry.once", map[string]string{"event": "retry"}))

	select {
	case <-processed:
	case <-time.After(8 * time.Second):
		t.Fatal("timed out waiting for redelivery")
	}

	assert.GreaterOrEqual(t, attempts.Load(), int32(2))
}

func TestNATSBroadcastAcrossTwoInstances(t *testing.T) {
	natsURL := setupNATSServer(t)

	eventBusA := newNATSEventBus(t, natsURL, "instance-a", 1, 1*time.Second)
	eventBusB := newNATSEventBus(t, natsURL, "instance-b", 1, 1*time.Second)
	t.Cleanup(func() {
		require.NoError(t, eventBusA.Close())
		require.NoError(t, eventBusB.Close())
	})

	subCtx, cancelSub := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancelSub()

	msgsA, err := eventBusA.Subscribe(subCtx, common.EventOrderReady)
	require.NoError(t, err)
	msgsB, err := eventBusB.Subscribe(subCtx, common.EventOrderReady)
	require.NoError(t, err)

	// Give subscriptions time to initialize before publishing.
	time.Sleep(200 * time.Millisecond)

	require.NoError(t, eventBusA.Publish(context.Background(), common.EventOrderReady, map[string]string{
		"order_id": "ord_broadcast_1",
	}))

	var gotA, gotB string
	select {
	case msg := <-msgsA:
		gotA = string(msg.Payload)
		msg.Ack()
	case <-time.After(6 * time.Second):
		t.Fatal("instance A did not receive broadcast event")
	}

	select {
	case msg := <-msgsB:
		gotB = string(msg.Payload)
		msg.Ack()
	case <-time.After(6 * time.Second):
		t.Fatal("instance B did not receive broadcast event")
	}

	assert.Contains(t, gotA, "ord_broadcast_1")
	assert.Contains(t, gotB, "ord_broadcast_1")
}

func newNATSEventBus(t *testing.T, natsURL, instanceID string, subscribers int, ackWait time.Duration) *events.EventBus {
	t.Helper()
	eventBus, err := events.NewEventBusWithConfig(events.Config{
		Backend:           "nats",
		NATSURL:           natsURL,
		NATSAutoProvision: true,
		NATSAckWait:       ackWait,
		NATSCloseTimeout:  5 * time.Second,
		NATSSubscribers:   subscribers,
		NATSInstanceID:    instanceID,
	})
	require.NoError(t, err)
	return eventBus
}

func newRouter(t *testing.T, closeTimeout time.Duration) *message.Router {
	t.Helper()
	router, err := message.NewRouter(message.RouterConfig{
		CloseTimeout: closeTimeout,
	}, watermill.NewStdLogger(false, false))
	require.NoError(t, err)
	return router
}

func runRouter(t *testing.T, router *message.Router) {
	t.Helper()
	runErr := make(chan error, 1)
	go func() {
		runErr <- router.Run(context.Background())
	}()

	select {
	case err := <-runErr:
		require.NoError(t, err)
	case <-router.Running():
	case <-time.After(5 * time.Second):
		t.Fatal("router did not report running state")
	}

	t.Cleanup(func() {
		_ = router.Close()
		select {
		case <-runErr:
		case <-time.After(3 * time.Second):
		}
	})
}

func setupNATSServer(t *testing.T) string {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping NATS integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "nats:2.11-alpine",
			ExposedPorts: []string{"4222/tcp"},
			Cmd:          []string{"-js", "-p", "4222"},
			WaitingFor: wait.ForListeningPort("4222/tcp").
				WithStartupTimeout(45 * time.Second),
		},
		Started: true,
	})
	if err != nil {
		t.Skipf("skipping NATS integration tests: cannot start container: %v", err)
	}

	t.Cleanup(func() {
		_ = container.Terminate(context.Background())
	})

	host, err := container.Host(ctx)
	require.NoError(t, err)
	mappedPort, err := container.MappedPort(ctx, "4222/tcp")
	require.NoError(t, err)

	return fmt.Sprintf("nats://%s:%s", host, mappedPort.Port())
}

type sseCaptureWriter struct {
	header http.Header
	writes chan []byte
	status int
}

func newSSECaptureWriter() *sseCaptureWriter {
	return &sseCaptureWriter{
		header: make(http.Header),
		writes: make(chan []byte, 32),
		status: http.StatusOK,
	}
}

func (w *sseCaptureWriter) Header() http.Header {
	return w.header
}

func (w *sseCaptureWriter) Write(data []byte) (int, error) {
	cp := append([]byte(nil), data...)
	select {
	case w.writes <- cp:
	default:
	}
	return len(data), nil
}

func (w *sseCaptureWriter) WriteHeader(statusCode int) {
	w.status = statusCode
}

func (w *sseCaptureWriter) Flush() {}

func (w *sseCaptureWriter) String() string {
	parts := make([]string, 0, len(w.writes))
	for {
		select {
		case msg := <-w.writes:
			parts = append(parts, string(msg))
		default:
			return strings.Join(parts, "")
		}
	}
}
