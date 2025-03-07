package repositories

import (
	"context"
	"database/sql"
	"fmt"

	uuid "github.com/satori/go.uuid"
	"github.com/vysogota0399/gophermart_billing/internal/logging"
	"github.com/vysogota0399/gophermart_billing/internal/models"
	"github.com/vysogota0399/gophermart_billing/internal/storage"
	"go.uber.org/zap"
)

type AccountingRepository struct {
	strg   AccountingStorage
	events TransactionEvents
	lg     *logging.ZapLogger
}

type AccountingStorage interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type TransactionEvents interface {
	NewTransactionProcessedTX(ctx context.Context, e *models.Transaction, tx *sql.Tx) error
}

func NewAccountingRepository(strg *storage.Storage, lg *logging.ZapLogger, e TransactionEvents) *AccountingRepository {
	return &AccountingRepository{strg: strg.DB, lg: lg, events: e}
}

func (rep *AccountingRepository) Create(ctx context.Context, in *models.Transaction) error {
	tx, err := rep.BeginTX(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := rep.CreateTX(ctx, in, tx); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("internal/repositories/accounting_repository commit tx error %w", err)
	}

	return nil
}

func (rep *AccountingRepository) CreateTX(ctx context.Context, in *models.Transaction, tx *sql.Tx) error {
	in.UUID = uuid.NewV4().String()
	rep.lg.DebugCtx(
		ctx,
		"create accounting transaction record",
		zap.Any("record", in),
	)

	row := tx.QueryRowContext(
		ctx,
		`
		  INSERT INTO accounting(uuid, order_number, account_id, amount, operation)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING created_at
		`,
		in.UUID, in.OrderNumber, in.AccountID, in.Amount.NanoBonuses(), in.Operation,
	)

	if err := row.Scan(&in.ProcessedAt); err != nil {
		return fmt.Errorf("internal/repositories/accounting_repository create accounting record error %w", err)
	}

	if err := rep.events.NewTransactionProcessedTX(ctx, in, tx); err != nil {
		return fmt.Errorf("internal/repositories/accounting_repository save transaction event %w", err)
	}

	return nil
}

func (rep *AccountingRepository) BalanceForUpdate(ctx context.Context, accountID int64, tx *sql.Tx) (int64, error) {
	debit, err := rep.totalAmmountForAccount(ctx, accountID, models.Debit, tx)
	if err != nil {
		return 0, fmt.Errorf("transactions_repository: calc debit error %w", err)
	}

	credit, err := rep.totalAmmountForAccount(ctx, accountID, models.Credit, tx)
	if err != nil {
		return 0, fmt.Errorf("transactions_repository: calc debit error %w", err)
	}

	return debit - credit, nil
}

func (rep *AccountingRepository) totalAmmountForAccount(
	ctx context.Context,
	accountID int64,
	operation string,
	tx *sql.Tx,
) (int64, error) {

	row := tx.QueryRowContext(
		ctx,
		`
			WITH rows AS ( 
				SELECT amount
				FROM accounting
				WHERE account_id = $1 AND operation = $2
        FOR UPDATE
		  )

			SELECT COALESCE(sum(rows.amount), 0) FROM rows;
			`,
		accountID, operation,
	)

	var result int64
	if err := row.Scan(&result); err != nil {
		return 0, fmt.Errorf("accounting_repository calc sum error %w", err)
	}

	return result, nil
}

func (rep *AccountingRepository) BeginTX(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return rep.strg.BeginTx(ctx, opts)
}

func (rep *AccountingRepository) CommitTX(tx *sql.Tx) error {
	return tx.Commit()
}
func (rep *AccountingRepository) RollbackTX(tx *sql.Tx) error {
	return tx.Rollback()
}
