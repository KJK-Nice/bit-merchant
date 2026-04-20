package events

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	watermillnats "github.com/ThreeDotsLabs/watermill-nats/v2/pkg/nats"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	natsgo "github.com/nats-io/nats.go"
)

const (
	backendMemory = "memory"
	backendNATS   = "nats"
)

var invalidInstanceIDChars = regexp.MustCompile(`[^a-zA-Z0-9_-]+`)

type closeable interface {
	Close() error
}

type natsConnectionCloser struct {
	conn *natsgo.Conn
}

func (c natsConnectionCloser) Close() error {
	if c.conn != nil {
		c.conn.Close()
	}
	return nil
}

// Config controls event bus backend selection and backend-specific settings.
type Config struct {
	Backend           string
	NATSURL           string
	NATSAutoProvision bool
	NATSAckWait       time.Duration
	NATSCloseTimeout  time.Duration
	NATSSubscribers   int
	NATSInstanceID    string
}

func (c Config) withDefaults() Config {
	if strings.TrimSpace(c.Backend) == "" {
		c.Backend = backendMemory
	}
	c.Backend = strings.ToLower(strings.TrimSpace(c.Backend))
	if strings.TrimSpace(c.NATSURL) == "" {
		c.NATSURL = "nats://localhost:4222"
	}
	if c.NATSAckWait <= 0 {
		c.NATSAckWait = 30 * time.Second
	}
	if c.NATSCloseTimeout <= 0 {
		c.NATSCloseTimeout = 30 * time.Second
	}
	if c.NATSSubscribers <= 0 {
		c.NATSSubscribers = 1
	}
	if strings.TrimSpace(c.NATSInstanceID) == "" {
		if hostname, err := os.Hostname(); err == nil {
			c.NATSInstanceID = hostname
		}
	}
	if strings.TrimSpace(c.NATSInstanceID) == "" {
		c.NATSInstanceID = "bitmerchant"
	}
	return c
}

// EventBus wraps Watermill pub/sub for domain events.
type EventBus struct {
	publisher  message.Publisher
	subscriber message.Subscriber
	closers    []closeable

	autoProvisionTopics bool
	jetstreamManager    natsgo.JetStreamContext
	provisionedTopics   map[string]struct{}
	provisionMu         sync.Mutex
}

// NewEventBus creates a default in-memory event bus.
func NewEventBus() *EventBus {
	eventBus, err := NewEventBusWithConfig(Config{})
	if err != nil {
		panic(err)
	}
	return eventBus
}

// NewEventBusWithConfig creates an event bus using the configured backend.
func NewEventBusWithConfig(cfg Config) (*EventBus, error) {
	cfg = cfg.withDefaults()

	switch cfg.Backend {
	case backendMemory:
		logger := watermill.NewStdLogger(false, false)
		pubSub := gochannel.NewGoChannel(gochannel.Config{}, logger)
		return &EventBus{
			publisher:  pubSub,
			subscriber: pubSub,
			closers:    uniqueClosers(pubSub),
		}, nil

	case backendNATS:
		wmLogger := watermill.NewStdLogger(false, false)
		instanceID := sanitizeInstanceID(cfg.NATSInstanceID)
		prefix := "bitmerchant_" + instanceID

		jetStreamCfg := watermillnats.JetStreamConfig{
			Disabled:      false,
			AutoProvision: false,
			TrackMsgId:    true,
			DurablePrefix: prefix,
			SubscribeOptions: []natsgo.SubOpt{
				natsgo.AckWait(cfg.NATSAckWait),
			},
		}

		publisher, err := watermillnats.NewPublisher(watermillnats.PublisherConfig{
			URL: cfg.NATSURL,
			JetStream: watermillnats.JetStreamConfig{
				Disabled:      false,
				AutoProvision: false,
				TrackMsgId:    true,
			},
		}, wmLogger)
		if err != nil {
			return nil, err
		}

		subscriber, err := watermillnats.NewSubscriber(watermillnats.SubscriberConfig{
			URL:              cfg.NATSURL,
			QueueGroupPrefix: prefix,
			SubscribersCount: cfg.NATSSubscribers,
			CloseTimeout:     cfg.NATSCloseTimeout,
			AckWaitTimeout:   cfg.NATSAckWait,
			JetStream:        jetStreamCfg,
		}, wmLogger)
		if err != nil {
			_ = publisher.Close()
			return nil, err
		}

		var provisionCloser closeable
		var jsManager natsgo.JetStreamContext
		if cfg.NATSAutoProvision {
			provisionConn, connErr := natsgo.Connect(cfg.NATSURL)
			if connErr != nil {
				_ = subscriber.Close()
				_ = publisher.Close()
				return nil, connErr
			}
			js, jsErr := provisionConn.JetStream()
			if jsErr != nil {
				provisionConn.Close()
				_ = subscriber.Close()
				_ = publisher.Close()
				return nil, jsErr
			}
			jsManager = js
			provisionCloser = natsConnectionCloser{conn: provisionConn}
		}

		return &EventBus{
			publisher:           publisher,
			subscriber:          subscriber,
			autoProvisionTopics: cfg.NATSAutoProvision,
			jetstreamManager:    jsManager,
			provisionedTopics:   make(map[string]struct{}),
			closers:             uniqueClosers(subscriber, publisher, provisionCloser),
		}, nil

	default:
		return nil, errors.New("unsupported event backend: " + cfg.Backend)
	}
}

func sanitizeInstanceID(instanceID string) string {
	sanitized := invalidInstanceIDChars.ReplaceAllString(strings.TrimSpace(instanceID), "_")
	sanitized = strings.Trim(sanitized, "_")
	if sanitized == "" {
		return "bitmerchant"
	}
	return sanitized
}

func uniqueClosers(closers ...closeable) []closeable {
	seen := make(map[closeable]struct{}, len(closers))
	unique := make([]closeable, 0, len(closers))

	for _, closer := range closers {
		if closer == nil {
			continue
		}
		if _, ok := seen[closer]; ok {
			continue
		}
		seen[closer] = struct{}{}
		unique = append(unique, closer)
	}

	return unique
}

// Publish publishes a domain event.
func (b *EventBus) Publish(ctx context.Context, topic string, event interface{}) error {
	if err := b.ensureTopic(topic); err != nil {
		return err
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	msg := message.NewMessage(watermill.NewUUID(), payload)
	return b.publisher.Publish(topic, msg)
}

// Subscribe subscribes to domain events.
func (b *EventBus) Subscribe(ctx context.Context, topic string) (<-chan *message.Message, error) {
	if err := b.ensureTopic(topic); err != nil {
		return nil, err
	}
	return b.subscriber.Subscribe(ctx, topic)
}

// Publisher returns the underlying Watermill publisher.
func (b *EventBus) Publisher() message.Publisher {
	return b.publisher
}

// Subscriber returns the underlying Watermill subscriber.
func (b *EventBus) Subscriber() message.Subscriber {
	return provisioningSubscriber{
		subscriber: b.subscriber,
		ensureTopic: func(topic string) error {
			return b.ensureTopic(topic)
		},
	}
}

// Close closes subscriber/publisher resources.
func (b *EventBus) Close() error {
	var err error
	for i := len(b.closers) - 1; i >= 0; i-- {
		err = errors.Join(err, b.closers[i].Close())
	}
	return err
}

func (b *EventBus) ensureTopic(topic string) error {
	if !b.autoProvisionTopics || b.jetstreamManager == nil {
		return nil
	}

	b.provisionMu.Lock()
	defer b.provisionMu.Unlock()

	if _, ok := b.provisionedTopics[topic]; ok {
		return nil
	}

	streamName := streamNameForTopic(topic)
	streamInfo, err := b.jetstreamManager.StreamInfo(streamName)
	if err == nil && streamInfo != nil {
		if contains(streamInfo.Config.Subjects, topic) {
			b.provisionedTopics[topic] = struct{}{}
			return nil
		}

		updatedConfig := streamInfo.Config
		updatedConfig.Subjects = append(updatedConfig.Subjects, topic)
		if _, err := b.jetstreamManager.UpdateStream(&updatedConfig); err != nil {
			return fmt.Errorf("update jetstream stream subjects for topic %q: %w", topic, err)
		}
		b.provisionedTopics[topic] = struct{}{}
		return nil
	}

	if _, err := b.jetstreamManager.AddStream(&natsgo.StreamConfig{
		Name:     streamName,
		Subjects: []string{topic},
	}); err != nil {
		return fmt.Errorf("provision jetstream stream for topic %q: %w", topic, err)
	}

	b.provisionedTopics[topic] = struct{}{}
	return nil
}

func streamNameForTopic(topic string) string {
	sanitized := invalidInstanceIDChars.ReplaceAllString(strings.TrimSpace(topic), "_")
	sanitized = strings.Trim(sanitized, "_")
	if sanitized == "" {
		return "bitmerchant_stream"
	}
	return "bm_" + sanitized
}

func contains(items []string, value string) bool {
	for _, item := range items {
		if item == value {
			return true
		}
	}
	return false
}

type provisioningSubscriber struct {
	subscriber  message.Subscriber
	ensureTopic func(string) error
}

func (p provisioningSubscriber) Subscribe(ctx context.Context, topic string) (<-chan *message.Message, error) {
	if err := p.ensureTopic(topic); err != nil {
		return nil, err
	}
	return p.subscriber.Subscribe(ctx, topic)
}

func (p provisioningSubscriber) Close() error {
	return p.subscriber.Close()
}
