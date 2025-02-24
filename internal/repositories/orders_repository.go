package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/vysogota0399/gophermart_billing/internal/logging"
	"github.com/vysogota0399/gophermart_billing/internal/models"
	"github.com/vysogota0399/gophermart_billing/internal/storage"
)

type OrdersRepository struct {
	strg   OrdersStorage
	events OrderEvents
	lg     *logging.ZapLogger
}

type OrdersStorage interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type OrderEvents interface {
	OrderCreated(ctx context.Context, in *models.Order, tx *sql.Tx) error
	OrderUpdated(ctx context.Context, in *models.Order, tx *sql.Tx) error
}

func NewOrdersRepository(strg *storage.Storage, lg *logging.ZapLogger, e OrderEvents) *OrdersRepository {
	return &OrdersRepository{strg: strg.DB, lg: lg, events: e}
}

var ErrDuplicateNumber = errors.New("order with this number already exists")

func (rep *OrdersRepository) CreateOrder(ctx context.Context, order *models.Order, tx *sql.Tx) error {
	_, err := tx.ExecContext(
		ctx,
		`
		  INSERT INTO ORDERS(uuid, state, number, account_id, uploaded_at)
			VALUES ($1, $2, $3, $4, $5)
		`,
		order.UUID, order.State, order.Number, order.AccountID, order.UploadedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
			return ErrDuplicateNumber
		}

		if err := tx.Rollback(); err != nil {
			return fmt.Errorf("internal/server/repositories/orders_repository save order rollback error %w", err)
		}

		return fmt.Errorf("internal/server/repositories/orders_repository save order error %w", err)
	}

	if err := rep.events.OrderCreated(ctx, order, tx); err != nil {
		if err := tx.Rollback(); err != nil {
			return fmt.Errorf("internal/server/repositories/orders_repository save event order_created rollback error %w", err)
		}

		return fmt.Errorf("internal/server/repositories/orders_repository save event order_created error %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("internal/server/repositories/orders_repository commit tx error %w", err)
	}
	return nil
}

func (rep *OrdersRepository) UpdateOrderState(ctx context.Context, in *models.Order, tx *sql.Tx) error {
	return rep.updateOrderState(ctx, in, tx)
}

func (rep *OrdersRepository) updateOrderState(ctx context.Context, in *models.Order, tx *sql.Tx) error {
	_, err := tx.ExecContext(
		ctx,
		`
			UPDATE orders
			SET state = $1
			WHERE uuid = $2
		`,
		in.State, in.UUID,
	)

	if err != nil {
		return fmt.Errorf("internal/server/repositories/orders_repository update order error %w", err)
	}

	if err := rep.events.OrderUpdated(ctx, in, tx); err != nil {
		return fmt.Errorf("internal/server/repositories/orders_repository save event order_updated error %w", err)
	}

	return nil
}

func (rep *OrdersRepository) FindByUUIDForUpdate(ctx context.Context, in *models.Order) (*sql.Tx, error) {
	tx, err := rep.strg.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("internal/server/repositories/orders_repository create tx for find for update order operation error %w", err)
	}

	row := tx.QueryRowContext(
		ctx,
		`
			SELECT state, number, uploaded_at, account_id
			FROM orders
			WHERE uuid = $1
			FOR UPDATE
		`,
		in.UUID,
	)

	if err := row.Scan(&in.State, &in.Number, &in.UploadedAt, &in.AccountID); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("internal/server/repositories/orders_repository find order for update operation error %w", err)
	}

	return tx, nil
}

func (rep *OrdersRepository) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return rep.strg.BeginTx(ctx, opts)
}
