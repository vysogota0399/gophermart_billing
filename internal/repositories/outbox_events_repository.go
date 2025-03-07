package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/vysogota0399/gophermart_billing/internal/logging"
	"github.com/vysogota0399/gophermart_billing/internal/models"
	"github.com/vysogota0399/gophermart_billing/internal/storage"
)

var OrderCreatedEventName = "order_created"
var OrderUpdatedEventName = "order_updated"
var TransactionProcessedEventName = "transaction_processed"

type OutboxEventsRepository struct {
	strg EventsStorage
	lg   *logging.ZapLogger
}

type EventsStorage interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func NewOutboxEventsRepository(strg *storage.Storage, lg *logging.ZapLogger) *OutboxEventsRepository {
	return &OutboxEventsRepository{strg: strg.DB, lg: lg}
}

func (rep *OutboxEventsRepository) OrderCreated(ctx context.Context, order *models.Order, tx *sql.Tx) error {
	var message []byte

	event_uuid := uuid.NewV4().String()
	e := &models.Meta{
		UUID:            event_uuid,
		OrderNumber:     order.Number,
		OrderUploadedAt: order.UploadedAt.Format(time.RFC3339Nano),
		OrderUUID:       order.UUID,
		OrderState:      order.State,
		OrderAccountID:  order.AccountID,
	}

	message, err := json.Marshal(e)
	if err != nil {
		return fmt.Errorf("internal/repositories/outbox_events_repository marshal event to json error %w", err)
	}

	query := `
			INSERT INTO outbox_events(uuid, name, message)
			VALUES ($1, $2, $3)
		`
	_, err = rep.execContext(ctx, tx, query, e.UUID, OrderCreatedEventName, message)
	if err != nil {
		return fmt.Errorf("internal/repositories/outbox_events_repository save order created event error %w", err)
	}

	return nil
}

func (rep *OutboxEventsRepository) OrderUpdated(ctx context.Context, order *models.Order, tx *sql.Tx) error {
	var message []byte

	e := &models.Meta{
		UUID:       uuid.NewV4().String(),
		OrderUUID:  order.UUID,
		OrderState: order.State,
	}

	message, err := json.Marshal(e)
	if err != nil {
		return fmt.Errorf("internal/repositories/outbox_events_repository marshal event to json error %w", err)
	}

	query := `
			INSERT INTO outbox_events(uuid, name, message)
			VALUES ($1, $2, $3)
		`

	_, err = rep.execContext(ctx, tx, query, e.UUID, OrderUpdatedEventName, message)

	if err != nil {
		return fmt.Errorf("internal/repositories/outbox_events_repository save order updated event error %w", err)
	}

	return nil
}

func (rep *OutboxEventsRepository) NewTransactionProcessedTX(ctx context.Context, in *models.Transaction, tx *sql.Tx) error {
	var message []byte

	meta := &models.TransactionEventMeta{
		UUID:            uuid.NewV4().String(),
		TransactionUUID: in.UUID,
		OrderNumber:     in.OrderNumber,
		AccountID:       in.AccountID,
		AmountUnits:     in.Amount.Money.Units,
		AmountNanos:     in.Amount.Money.Nanos,
		Operation:       in.Operation,
		ProcessedAt:     in.ProcessedAt,
	}

	message, err := json.Marshal(meta)
	if err != nil {
		return fmt.Errorf("internal/repositories/outbox_events_repository marshal event to json error %w", err)
	}

	query := `
			INSERT INTO outbox_events(uuid, name, message)
			VALUES ($1, $2, $3)
		`

	_, err = rep.execContext(ctx, tx, query, meta.UUID, TransactionProcessedEventName, message)

	if err != nil {
		return fmt.Errorf("internal/repositories/outbox_events_repository save transaction_processedevent error %w", err)
	}

	return nil
}

func (rep *OutboxEventsRepository) Send(ctx context.Context, e *models.OrderEvent, tx *sql.Tx) error {
	if _, err := tx.ExecContext(
		ctx,
		`
			UPDATE outbox_events
			SET SEND = true
			WHERE uuid = $1
		`,
		e.UUID,
	); err != nil {
		return fmt.Errorf("outbox_events_repository: save event send error %w", err)
	}

	return nil
}

func (rep *OutboxEventsRepository) SendTransactionEvent(ctx context.Context, e *models.TransactionEvent, tx *sql.Tx) error {
	if _, err := tx.ExecContext(
		ctx,
		`
			UPDATE outbox_events
			SET send = true
			WHERE uuid = $1
		`,
		e.UUID,
	); err != nil {
		return fmt.Errorf("outbox_events_repository: save event send error %w", err)
	}

	return nil
}

func (rep *OutboxEventsRepository) ReserveNewTransactionEvent(ctx context.Context, eventName string, tx *sql.Tx) (*models.TransactionEvent, error) {
	e := &models.TransactionEvent{}
	row := tx.QueryRowContext(
		ctx,
		`
			SELECT uuid, name, message
			FROM outbox_events
			WHERE name = $1 AND send = false
			ORDER BY created_at ASC
			FOR UPDATE SKIP LOCKED
			LIMIT 1
		`,
		eventName)

	if err := row.Scan(&e.UUID, &e.Name, &e.Meta); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("outbox_events_repository: select event error %w", err)
	}

	return e, nil
}

func (rep *OutboxEventsRepository) ReserveNewOrderEvent(ctx context.Context, eventName string, tx *sql.Tx) (*models.OrderEvent, error) {
	e := &models.OrderEvent{}
	row := tx.QueryRowContext(
		ctx,
		`
			SELECT uuid, name, message
			FROM outbox_events
			WHERE name = $1 AND send = false
			ORDER BY created_at ASC
			FOR UPDATE SKIP LOCKED 
			LIMIT 1
		`,
		eventName)

	if err := row.Scan(&e.UUID, &e.Name, &e.Meta); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("outbox_events_repository: select event error %w", err)
	}

	return e, nil
}

type executor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func (rep *OutboxEventsRepository) execContext(
	ctx context.Context,
	ex executor,
	query string,
	args ...any,
) (sql.Result, error) {
	return ex.ExecContext(ctx, query, args...)
}

func (rep *OutboxEventsRepository) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return rep.strg.BeginTx(ctx, opts)
}

func (rep *OutboxEventsRepository) CommitTX(tx *sql.Tx) error {
	return tx.Commit()
}
func (rep *OutboxEventsRepository) RollbackTX(tx *sql.Tx) error {
	return tx.Rollback()
}
