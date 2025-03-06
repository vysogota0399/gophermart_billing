package commands

import (
	"context"
	"errors"

	"github.com/vysogota0399/gophermart_billing/internal/logging"
	"github.com/vysogota0399/gophermart_billing/internal/models"
	"github.com/vysogota0399/gophermart_billing/internal/repositories"
	"github.com/vysogota0399/gophermart_protos/gen/commands/create_order"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type CreateOrderCommand struct {
	create_order.UnimplementedCreateOrderServiceServer

	lg  *logging.ZapLogger
	srv CreateOrderService
}

type CreateOrderService interface {
	Call(context.Context, *create_order.CreateNewOrderParams) (*models.Order, error)
}

func NewCreateOrderCommand(srv CreateOrderService, lg *logging.ZapLogger) *CreateOrderCommand {
	return &CreateOrderCommand{srv: srv, lg: lg}
}

var ErrOrderAlreadtExists = status.Errorf(codes.AlreadyExists, "order already exists")

func (cmd *CreateOrderCommand) Create(ctx context.Context, order *create_order.CreateNewOrderParams) (*emptypb.Empty, error) {
	ctx = cmd.lg.WithContextFields(ctx, zap.String("actor", "order_service_command"))

	if _, err := cmd.srv.Call(ctx, order); err != nil {
		if errors.Is(err, repositories.ErrDuplicateNumber) {
			cmd.lg.ErrorCtx(ctx, "create order failed - order number duplicate", zap.Error(err))
			return nil, ErrOrderAlreadtExists
		}

		cmd.lg.ErrorCtx(ctx, "create order failed - internal error", zap.Error(err))
		return nil, err
	}

	return nil, nil
}
