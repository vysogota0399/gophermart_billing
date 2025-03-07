package services

import (
	"context"
	"fmt"

	"github.com/vysogota0399/gophermart_billing/internal/logging"
	"github.com/vysogota0399/gophermart_billing/internal/models"
	"github.com/vysogota0399/gophermart_billing/internal/server/entities"
	"github.com/vysogota0399/gophermart_protos/gen/commands/create_order"
)

type CreateOrderService struct {
	lg       *logging.ZapLogger
	orderFsm OrderStateMachineContainer
}

type OrderStateMachineContainer interface {
	Create(ctx context.Context, opt entities.OrderFsmOption) (*models.Order, error)
}

func NewCreateOrderService(lg *logging.ZapLogger, orderFsm OrderStateMachineContainer) *CreateOrderService {
	return &CreateOrderService{lg: lg, orderFsm: orderFsm}
}

func (srv *CreateOrderService) Call(ctx context.Context, in *create_order.CreateNewOrderParams) (*models.Order, error) {
	order, err := srv.orderFsm.Create(
		ctx,
		entities.OrderFsmOption{
			Order: &models.Order{
				UUID:       in.Uuid.Value,
				Number:     in.Number,
				UploadedAt: in.UploadedAt.AsTime(),
				AccountID:  in.Account.Id,
			},
		},
	)

	if err != nil {
		return nil, fmt.Errorf("services/create_order_service create order error %w", err)
	}

	return order, nil
}
