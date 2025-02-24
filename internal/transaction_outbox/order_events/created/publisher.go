package created

import (
	"context"
	"fmt"

	"github.com/segmentio/kafka-go"
	"github.com/vysogota0399/gophermart_billing/internal/config"
	"github.com/vysogota0399/gophermart_billing/internal/logging"
	"github.com/vysogota0399/gophermart_billing/internal/models"
	"github.com/vysogota0399/gophermart_billing/internal/transaction_outbox"
	"github.com/vysogota0399/gophermart_protos/gen/events"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type Publisher struct {
	lg     *logging.ZapLogger
	writer *kafka.Writer
}

func NewPublisher(
	lg *logging.ZapLogger,
	cfg *transaction_outbox.Config,
	globalCFG *config.Config,
	errLogger *logging.KafkaErrorLogger,
	logger *logging.KafkaLogger,
) *Publisher {
	w := kafka.NewWriter(
		kafka.WriterConfig{
			Brokers:      globalCFG.KafkaBrokers,
			Topic:        cfg.KafkaOrderCreatedTopic,
			RequiredAcks: 0,
			Logger:       logger,
			ErrorLogger:  errLogger,
			Balancer:     &kafka.Hash{},
		},
	)

	return &Publisher{lg: lg, writer: w}
}

func (p *Publisher) Publish(ctx context.Context, e *models.OrderEvent) error {
	order := events.OrderCreated{
		EventUuid:  e.Meta.UUID,
		Uuid:       e.Meta.OrderUUID,
		Number:     e.Meta.OrderNumber,
		State:      e.Meta.OrderState,
		UploadedAt: e.Meta.OrderUploadedAt,
		AccountId:  e.Meta.OrderAccountID,
	}

	event, err := proto.Marshal(&order)
	if err != nil {
		return fmt.Errorf("order_events/created/publisher: marashal failed  %w", err)
	}

	payload := kafka.Message{
		Key:   []byte(e.Meta.OrderNumber),
		Value: event,
	}

	p.lg.DebugCtx(
		ctx,
		"publish message to topic",
		zap.String("topic", p.writer.Topic),
		zap.Any("message", e),
	)

	if err := p.writer.WriteMessages(
		ctx,
		payload,
	); err != nil {
		return fmt.Errorf("order_events/created/publisher: write message to toppic failed %w", err)
	}

	return nil
}
