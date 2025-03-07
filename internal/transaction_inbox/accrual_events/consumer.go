package accrual_events

import (
	"context"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/vysogota0399/gophermart_billing/internal/config"
	"github.com/vysogota0399/gophermart_billing/internal/logging"
	"github.com/vysogota0399/gophermart_billing/internal/models"
	"github.com/vysogota0399/gophermart_billing/internal/repositories"
	"github.com/vysogota0399/gophermart_billing/internal/transaction_inbox"
	"github.com/vysogota0399/gophermart_protos/gen/events"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type Consumer struct {
	lg        *logging.ZapLogger
	reader    *kafka.Reader
	events    InboxEventsRepository
	cancaller context.CancelFunc
	globalCtx context.Context

	accrualFailed   AccrualFailedHandler
	accrualFinished AccrualFinishedHandler
	accrualStarted  AccrualStartedHandler
}

type AccrualFailedHandler interface {
	Call(ctx context.Context, event *events.AccrualFailedEvent)
}

type AccrualFinishedHandler interface {
	Call(ctx context.Context, event *events.AccrualFinishedEvent)
}

type AccrualStartedHandler interface {
	Call(ctx context.Context, event *events.AccrualStartedEvent)
}

type InboxEventsRepository interface {
	SaveAccrualEvent(ctx context.Context, in *models.AccrualEvent) error
}

func NewConsumer(
	lc fx.Lifecycle,
	lg *logging.ZapLogger,
	cfg *transaction_inbox.Config,
	globalCFG *config.Config,
	errLogger *logging.KafkaErrorLogger,
	logger *logging.KafkaLogger,
	events InboxEventsRepository,
	accrualFailed AccrualFailedHandler,
	accrualFinished AccrualFinishedHandler,
	accrualStarted AccrualStartedHandler,
) *Consumer {
	r := kafka.NewReader(kafka.ReaderConfig{
		GroupID:                cfg.KafkaAccrualsGroupID,
		PartitionWatchInterval: time.Duration(cfg.KafkaAccrualsPartitionWatchInterval) * time.Millisecond,
		Brokers:                globalCFG.KafkaBrokers,
		Topic:                  cfg.KafkaAccrualsTopic,
		MinBytes:               10e2, // 1KB
		MaxBytes:               10e6, // 10MB
		ErrorLogger:            errLogger,
		MaxWait:                time.Duration(cfg.KafkaAccrualsdMaxWaitInterval) * time.Millisecond,
		Logger:                 logger,
	})

	cns := &Consumer{
		lg:              lg,
		reader:          r,
		events:          events,
		accrualFailed:   accrualFailed,
		accrualFinished: accrualFinished,
		accrualStarted:  accrualStarted,
	}

	lc.Append(
		fx.Hook{
			OnStart: func(ctx context.Context) error {
				for {
					lg.InfoCtx(
						ctx,
						"Consumer started",
						zap.Any("global_config", globalCFG),
						zap.Any("daemons_config", cfg),
					)
					go cns.consume()
					return nil
				}
			},
			OnStop: func(ctx context.Context) error {
				return cns.reader.Close()
			},
		},
	)

	return cns
}

func (cns *Consumer) consume() {
	ctx, cancel := context.WithCancel(context.Background())
	cns.globalCtx = ctx
	cns.cancaller = cancel

	for {
		select {
		case <-ctx.Done():
			cns.lg.DebugCtx(ctx, "consumer graceful shutdown")
			return
		default:
			if err := cns.processMessage(cns.globalCtx); err != nil {
				cns.lg.ErrorCtx(ctx, "accrual_events/consumer: fetch message error", zap.Error(err))
			}
		}
	}
}

func (cns *Consumer) processMessage(ctx context.Context) error {
	m, err := cns.reader.FetchMessage(cns.globalCtx)
	if err != nil {
		return fmt.Errorf("accrual_events/consumer: fetch message error %w", err)
	}

	handler, err := cns.saveEvent(ctx, &m)
	if err != nil {
		return err
	}

	go handler(ctx)

	return nil
}

type eventHandler func(context.Context)

func (cns *Consumer) saveEvent(ctx context.Context, m *kafka.Message) (eventHandler, error) {
	payload := &events.AccrualProcessed{}

	if err := proto.Unmarshal(m.Value, payload); err != nil {
		return nil, fmt.Errorf("accrual_events/consumer: unmarshal message error %w", err)
	}

	var handler eventHandler
	var event *models.AccrualEvent

	switch message := payload.Event.(type) {
	case *events.AccrualProcessed_FailedEvent:
		event = &models.AccrualEvent{
			UUID:  message.FailedEvent.EventUuid.Value,
			Name:  repositories.AccrualFailedEventName,
			State: models.AccrualNewState,
			Meta: &models.AccrualEventMeta{
				EventUUID:   message.FailedEvent.EventUuid.Value,
				OrderUUID:   message.FailedEvent.OrderUuid.Value,
				OrderNumber: message.FailedEvent.OrderNumber,
			},
		}
		handler = func(ctx context.Context) {
			cns.accrualFailed.Call(ctx, message.FailedEvent)
		}
	case *events.AccrualProcessed_FinishedEvent:
		event = &models.AccrualEvent{
			UUID:  message.FinishedEvent.EventUuid.Value,
			Name:  repositories.AccrualFinishedEventName,
			State: models.AccrualNewState,
			Meta: &models.AccrualEventMeta{
				EventUUID:   message.FinishedEvent.EventUuid.Value,
				OrderUUID:   message.FinishedEvent.OrderUuid.Value,
				OrderNumber: message.FinishedEvent.OrderNumber,
				Amount:      message.FinishedEvent.Amount.Units,
			},
		}

		handler = func(ctx context.Context) {
			cns.accrualFinished.Call(ctx, message.FinishedEvent)
		}
	case *events.AccrualProcessed_StartedEvent:
		event = &models.AccrualEvent{
			UUID:  message.StartedEvent.EventUuid.Value,
			Name:  repositories.AccrualStartedEventName,
			State: models.AccrualNewState,
			Meta: &models.AccrualEventMeta{
				EventUUID:   message.StartedEvent.EventUuid.Value,
				OrderUUID:   message.StartedEvent.OrderUuid.Value,
				OrderNumber: message.StartedEvent.OrderNumber,
			},
		}

		handler = func(ctx context.Context) {
			cns.accrualStarted.Call(ctx, message.StartedEvent)
		}
	}

	if handler == nil {
		return nil, fmt.Errorf("accrual_events/consumer: consumed underfined event type")
	}

	cns.lg.InfoCtx(ctx, "consumed message", zap.Any("message", event))
	ctx = cns.lg.WithContextFields(
		ctx,
		zap.String("event_uuid", event.Meta.EventUUID),
		zap.String("order_number", event.Meta.OrderNumber),
	)

	if err := cns.events.SaveAccrualEvent(ctx, event); err != nil {
		return nil, fmt.Errorf("accrual_events/consumer: save message error %w", err)
	}

	return handler, nil
}
