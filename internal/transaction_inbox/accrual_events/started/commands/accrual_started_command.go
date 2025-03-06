package commands

import (
	"context"
	"fmt"

	"github.com/vysogota0399/gophermart_billing/internal/logging"
	"github.com/vysogota0399/gophermart_billing/internal/models"
	"github.com/vysogota0399/gophermart_billing/internal/server/entities"
	events "github.com/vysogota0399/gophermart_protos/gen/events"
	"go.uber.org/zap"
)

type AccrualStartedCommand struct {
	lg       *logging.ZapLogger
	orderFsm OrderFSM
}

type OrderFSM interface {
	Start(ctx context.Context, opt entities.OrderFsmOption) (*models.Order, error)
}

func NewAccrualStartedCommand(orderFsm OrderFSM, lg *logging.ZapLogger) *AccrualStartedCommand {
	return &AccrualStartedCommand{lg: lg, orderFsm: orderFsm}
}

func (cmd *AccrualStartedCommand) Call(ctx context.Context, acc *events.AccrualStartedEvent) (*models.Order, error) {
	ctx = cmd.lg.WithContextFields(ctx, zap.String("actor", "accrual_started_command"))
	order, err := cmd.orderFsm.Start(ctx, entities.OrderFsmOption{UUID: acc.OrderUuid.Value})
	if err != nil {
		return nil, fmt.Errorf("accruals/started/Accrual_started_command set order started error %w", err)
	}

	return order, nil
}
