package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/vysogota0399/gophermart_billing/internal/logging"
	"github.com/vysogota0399/gophermart_billing/internal/models"
	"github.com/vysogota0399/gophermart_protos/gen/commands/withdraw"
	"go.uber.org/zap"
)

type WithdrawService struct {
	lg         *logging.ZapLogger
	rep        WithdrawRepository
	retryCount int64
}

type WithdrawRepository interface {
	CreateTX(context.Context, *models.Transaction, *sql.Tx) error
	BalanceForUpdate(context.Context, int64, *sql.Tx) (int64, error)
	BeginTX(context.Context, *sql.TxOptions) (*sql.Tx, error)
	CommitTX(*sql.Tx) error
	RollbackTX(*sql.Tx) error
}

func NewWithdrawService(rep WithdrawRepository, lg *logging.ZapLogger) *WithdrawService {
	return &WithdrawService{rep: rep, lg: lg, retryCount: 2}
}

var ErrNotEnoughFunds error = errors.New("internal/server/services/withreaw_service not enough funds error")
var ErrAccountDeadlock error = errors.New("internal/server/services/withreaw_service lock_not_available error")

func (srv *WithdrawService) Call(ctx context.Context, wd *withdraw.WithdrawParams) error {
	process := func(
		ctx context.Context,
		wd *withdraw.WithdrawParams,
	) error {
		rep := srv.rep
		tx, err := rep.BeginTX(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
		if err != nil {
			return fmt.Errorf("internal/server/services/withreaw_service withraw open tx error %w", err)
		}
		defer rep.RollbackTX(tx)

		balance, err := rep.BalanceForUpdate(ctx, wd.Account.Id, tx)
		if err != nil {
			return fmt.Errorf("internal/server/services/withreaw_service withraw calculate blaance error %w", err)
		}

		withdraw_amount := int64(wd.Amount.GetUnits() * 100)
		srv.lg.DebugCtx(
			ctx,
			"withdraw debug information",
			zap.String("result", "failed"),
			zap.Int64("current_balance", balance),
			zap.Int64("withdtaw_amount", withdraw_amount),
			zap.Int64("account_id", wd.Account.Id),
		)

		if balance < withdraw_amount {
			return ErrNotEnoughFunds
		}

		credit := &models.Transaction{
			Amount:      withdraw_amount,
			Operation:   models.Credit,
			OrderNumber: wd.OrderNumber,
			AccountID:   wd.Account.Id,
			ProcessedAt: time.Now().Local(),
		}

		if err := srv.rep.CreateTX(ctx, credit, tx); err != nil {
			return fmt.Errorf("internal/server/services/withreaw_service withraw error %w", err)
		}

		return rep.CommitTX(tx)
	}

	var pgErr *pgconn.PgError

	for i := 0; i < int(srv.retryCount); i++ {
		err := process(ctx, wd)
		if err == nil {
			return nil
		}

		if errors.As(err, &pgErr) && pgerrcode.IsTransactionRollback(pgErr.Code) {
			continue
		} else {
			return err
		}
	}

	return pgErr
}
