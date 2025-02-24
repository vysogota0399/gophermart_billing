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

type AccrualFailedCommand struct {
	lg       *logging.ZapLogger
	orderFsm AccrualFailder
}

type AccrualFailder interface {
	CostAccrualFailed(ctx context.Context, opt entities.OrderFsmOption) (*models.Order, error)
}

func NewAccrualFailedCommand(orderFsm AccrualFailder, lg *logging.ZapLogger) *AccrualFailedCommand {
	return &AccrualFailedCommand{lg: lg, orderFsm: orderFsm}
}

func (cmd *AccrualFailedCommand) Call(ctx context.Context, acc *events.FailedEvent) (*models.Order, error) {
	ctx = cmd.lg.WithContextFields(ctx, zap.String("actor", "accural_failed_command"))
	cmd.lg.DebugCtx(ctx, "cost accural failed", zap.Any("order_uuid", acc.OrderUuid))

	order, err := cmd.orderFsm.CostAccrualFailed(
		ctx,
		entities.OrderFsmOption{
			UUID: acc.OrderUuid,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("accural_failed_command set order accural error %w", err)
	}

	return order, nil
}
