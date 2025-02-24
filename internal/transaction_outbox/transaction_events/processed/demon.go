package processed

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/vysogota0399/gophermart_billing/internal/logging"
	"github.com/vysogota0399/gophermart_billing/internal/models"
	"github.com/vysogota0399/gophermart_billing/internal/repositories"
	"github.com/vysogota0399/gophermart_billing/internal/transaction_outbox"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Daemon struct {
	lg           *logging.ZapLogger
	pollInterval time.Duration
	workersCount int64

	cfg *transaction_outbox.Config

	cancaller context.CancelFunc
	globalCtx context.Context
	events    TransactionProcessedEventsRepository

	publisher TransactionProcessedPublisher
}

type TransactionProcessedPublisher interface {
	Publish(ctx context.Context, e *models.TransactionEvent) error
}

type TransactionProcessedEventsRepository interface {
	SendTransactionEvent(ctx context.Context, e *models.TransactionEvent, tx *sql.Tx) error
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	CommitTX(*sql.Tx) error
	RollbackTX(*sql.Tx) error
	ReserveNewTransactionEvent(ctx context.Context, eventName string, tx *sql.Tx) (*models.TransactionEvent, error)
}

func NewDaemon(lc fx.Lifecycle, publisher TransactionProcessedPublisher, events TransactionProcessedEventsRepository, lg *logging.ZapLogger, cfg *transaction_outbox.Config) *Daemon {
	dmn := &Daemon{
		lg:           lg,
		pollInterval: time.Duration(cfg.AccountingCreatedPollInterval) * time.Millisecond,
		workersCount: cfg.WorkersCount,
		publisher:    publisher,
		events:       events,
		cfg:          cfg,
	}
	lc.Append(
		fx.Hook{
			OnStart: func(ctx context.Context) error {
				dmn.Start()
				return nil
			},
			OnStop: func(ctx context.Context) error {
				dmn.cancaller()
				return nil
			},
		},
	)

	return dmn
}

func (dmn *Daemon) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	dmn.cancaller = cancel
	dmn.globalCtx = dmn.lg.WithContextFields(ctx, zap.String("name", "order_events_daemon"))

	dmn.lg.InfoCtx(
		ctx,
		fmt.Sprintf("start processong %s events", repositories.OrderCreatedEventName),
		zap.Any("config", dmn.cfg),
	)

	for i := 0; i < int(dmn.workersCount); i++ {
		wctx := dmn.lg.WithContextFields(ctx, zap.Int("worker_id", i))
		go func() {
			ticker := time.NewTicker(dmn.pollInterval)

			for {
				select {
				case <-wctx.Done():
					dmn.lg.InfoCtx(wctx, "daemon worker graceful shutdown")
					return
				case <-ticker.C:
					if err := dmn.processEvent(wctx); err != nil {
						dmn.lg.ErrorCtx(ctx, "process event order created error", zap.Error(err))
					}
				}
			}
		}()
	}
}

func (dmn *Daemon) processEvent(ctx context.Context) error {
	tx, err := dmn.events.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin tx error %w", err)
	}
	defer dmn.events.RollbackTX(tx)

	e, err := dmn.events.ReserveNewTransactionEvent(ctx, repositories.TransactionProcessedEventName, tx)
	if err != nil {
		return fmt.Errorf("find first unpublished, process failed %w", err)
	}

	if e == nil {
		if err := dmn.events.CommitTX(tx); err != nil {
			return fmt.Errorf("commit error %w", err)
		}
		return nil
	}

	if err := dmn.publisher.Publish(ctx, e); err != nil {
		return fmt.Errorf("publish models.Event, process models.Event failed %w", err)
	}

	if err := dmn.events.SendTransactionEvent(ctx, e, tx); err != nil {
		return fmt.Errorf("update event as send, process models.Event failed %w", err)
	}

	if err := dmn.events.CommitTX(tx); err != nil {
		return fmt.Errorf("commit error %w", err)
	}

	return nil
}
