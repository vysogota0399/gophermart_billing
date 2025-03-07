package commands

import (
	"context"
	"errors"

	"github.com/vysogota0399/gophermart_billing/internal/logging"
	"github.com/vysogota0399/gophermart_billing/internal/server/services"
	"github.com/vysogota0399/gophermart_protos/gen/commands/withdraw"
	"go.uber.org/zap"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type WithdrawCommand struct {
	withdraw.UnimplementedWithdrawServiceServer
	lg  *logging.ZapLogger
	srv WithdrawService
}

func NewCreateWithdrawCommand(srv WithdrawService, lg *logging.ZapLogger) *WithdrawCommand {
	return &WithdrawCommand{lg: lg, srv: srv}
}

type WithdrawService interface {
	Call(context.Context, *withdraw.DoWithdrawParams) error
}

var ErrWithdrawInternalError = status.Error(codes.Internal, "internal error")

func (cmd *WithdrawCommand) DoWithdraw(ctx context.Context, wp *withdraw.DoWithdrawParams) (*withdraw.DoWithdrawParams, error) {
	ctx = cmd.lg.WithContextFields(ctx, zap.String("actor", "order_service_command"))
	if err := cmd.srv.Call(ctx, wp); err != nil {
		if errors.Is(err, services.ErrNotEnoughFunds) {
			st := status.New(codes.Aborted, "withdraw failed")
			ds, err := st.WithDetails(
				&errdetails.BadRequest{
					FieldViolations: []*errdetails.BadRequest_FieldViolation{{
						Field:       "Amount",
						Description: "there are insufficient funds on the balance",
					}},
				},
			)

			if err != nil {
				cmd.lg.ErrorCtx(ctx, "add datails to error failed", zap.Error(err))
				return nil, st.Err()
			}

			cmd.lg.ErrorCtx(ctx, "withdrawal failed - there are insufficient funds on the balance", zap.Error(err))
			return nil, ds.Err()
		}

		cmd.lg.ErrorCtx(ctx, "withdrawal failed - internal error", zap.Error(err))
		return nil, errors.Join(ErrWithdrawInternalError, err)
	}

	return nil, nil
}
