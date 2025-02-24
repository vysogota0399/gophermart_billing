package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/vysogota0399/gophermart_billing/internal/logging"
	"github.com/vysogota0399/gophermart_billing/internal/models"
	"github.com/vysogota0399/gophermart_billing/internal/server/entities"
	events "github.com/vysogota0399/gophermart_protos/gen/events"
	"go.uber.org/zap"
)

type CreateDebitCommand struct {
	lg         *logging.ZapLogger
	orderFsm   AccrualCreator
	accounting DebitCreator
}

type AccrualCreator interface {
	CostAccrualed(ctx context.Context, amount int64, opt entities.OrderFsmOption) (*models.Order, error)
}

type DebitCreator interface {
	Create(ctx context.Context, in *models.Transaction) error
}

func NewCreateDebitCommand(orderFsm AccrualCreator, accounting DebitCreator, lg *logging.ZapLogger) *CreateDebitCommand {
	return &CreateDebitCommand{lg: lg, orderFsm: orderFsm, accounting: accounting}
}

func (cmd *CreateDebitCommand) Call(ctx context.Context, acc *events.FinishedEvent) (*models.Order, error) {
	ctx = cmd.lg.WithContextFields(ctx, zap.String("actor", "set_accrual_command"))
	cmd.lg.DebugCtx(ctx, "update order state")

	order, err := cmd.orderFsm.CostAccrualed(
		ctx,
		acc.Amount,
		entities.OrderFsmOption{
			UUID: acc.OrderUuid,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("set_accrual_command: order accrual error %w", err)
	}

	cmd.lg.DebugCtx(ctx, "update order state - finished")
	cmd.lg.DebugCtx(ctx, "create transaction")

	debit := &models.Transaction{
		Amount:      acc.Amount,
		Operation:   models.Debit,
		OrderNumber: order.Number,
		AccountID:   order.AccountID,
		ProcessedAt: time.Now().Local(),
	}

	if err := cmd.accounting.Create(ctx, debit); err != nil {
		return nil, fmt.Errorf("set_accrual_command: create transaction error %w", err)
	}

	cmd.lg.DebugCtx(ctx, "transaction created", zap.Any("transaction", debit))

	return order, nil
}
