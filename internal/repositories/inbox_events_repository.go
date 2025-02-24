package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/vysogota0399/gophermart_billing/internal/logging"
	"github.com/vysogota0399/gophermart_billing/internal/models"
	"github.com/vysogota0399/gophermart_billing/internal/storage"
)

var AccrualFailedEventName = "accrual_failed"
var AccrualFinishedEventName = "accrual_finished"
var AccrualStartedEventName = "accrual_started"

type InboxEventsRepository struct {
	strg InboxEventsStorage
	lg   *logging.ZapLogger
}

type InboxEventsStorage interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func NewInboxEventsRepository(strg *storage.Storage, lg *logging.ZapLogger) *InboxEventsRepository {
	return &InboxEventsRepository{strg: strg.DB, lg: lg}
}

func (rep *InboxEventsRepository) ReserveEvent(ctx context.Context, eventName string) (*models.AccrualEvent, error) {
	tx, err := rep.strg.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("inbox_events_repository: create tx error %w", err)
	}
	defer tx.Rollback()

	e := &models.AccrualEvent{}
	row := tx.QueryRow(`
											SELECT uuid, name, state, message
											FROM inbox_events
											WHERE name = $1 AND state = $2
											ORDER BY created_at ASC
											FOR UPDATE SKIP LOCKED
											limit 1
										`,
		eventName, models.AccrualNewState)

	if err := row.Scan(&e.UUID, &e.Name, &e.State, &e.Meta); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("inbox_events_repository: select event error %w", err)
	}

	if err := rep.setStateTX(ctx, e.UUID, models.AccrualFailedState, tx); err != nil {
		return nil, fmt.Errorf("inbox_events_repository: set new state error %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("inbox_events_repository: commit tx error %w", err)
	}

	return e, nil
}

func (rep *InboxEventsRepository) SetState(ctx context.Context, uuid string, newState string, txs ...*sql.Tx) error {
	if len(txs) == 1 {
		return rep.setStateTX(ctx, uuid, newState, txs[0])
	}

	if _, err := rep.strg.ExecContext(ctx,
		`
																			UPDATE inbox_events
																			SET state = $1
																			WHERE uuid = $2
																		`,
		newState, uuid); err != nil {
		return fmt.Errorf("inbox_events_repository: update event state error %w", err)
	}

	return nil
}

func (rep *InboxEventsRepository) setStateTX(ctx context.Context, uuid string, newState string, tx *sql.Tx) error {
	if _, err := tx.ExecContext(ctx,
		`
																			UPDATE inbox_events
																			SET state = $1
																			WHERE uuid = $2
																		`,
		newState, uuid); err != nil {
		return fmt.Errorf("inbox_events_repository: update event state error %w", err)
	}

	return nil
}

func (rep *InboxEventsRepository) SaveAccrualEvent(ctx context.Context, in *models.AccrualEvent) error {
	if _, err := rep.strg.ExecContext(ctx,
		`
			INSERT INTO inbox_events (uuid, state, name, message)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT DO NOTHING
		`,
		in.UUID, in.State, in.Name, in.Meta); err != nil {
		return fmt.Errorf("inbox_events_repository: crate event error %w", err)
	}

	return nil
}
