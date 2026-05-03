package notification_test

import (
	"context"
	"errors"
	"testing"

	"bitmerchant/internal/infrastructure/logging"
	"bitmerchant/internal/notification"
)

type stubNotifier struct {
	name    string
	sendErr error
	calls   []notification.Notification
}

func (s *stubNotifier) Name() string { return s.name }
func (s *stubNotifier) Send(_ context.Context, n notification.Notification) error {
	s.calls = append(s.calls, n)
	return s.sendErr
}

func TestService_FanOut_BothReceive(t *testing.T) {
	a := &stubNotifier{name: "a"}
	b := &stubNotifier{name: "b"}
	svc := notification.NewService(logging.NewLogger(), a, b)

	n := notification.Notification{Title: "T", Body: "B"}
	svc.Send(context.Background(), n)

	if len(a.calls) != 1 || len(b.calls) != 1 {
		t.Fatalf("expected both notifiers to receive 1 call, got a=%d b=%d", len(a.calls), len(b.calls))
	}
}

func TestService_FanOut_OneErrorDoesNotBlockOther(t *testing.T) {
	a := &stubNotifier{name: "a", sendErr: errors.New("channel down")}
	b := &stubNotifier{name: "b"}
	svc := notification.NewService(logging.NewLogger(), a, b)

	svc.Send(context.Background(), notification.Notification{Title: "T"})

	if len(b.calls) != 1 {
		t.Fatalf("expected b to still receive notification after a errored, got %d calls", len(b.calls))
	}
}
