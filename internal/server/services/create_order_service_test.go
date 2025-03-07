package services

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/vysogota0399/gophermart_billing/internal/config"
	"github.com/vysogota0399/gophermart_billing/internal/logging"
	"github.com/vysogota0399/gophermart_billing/internal/models"
	sm "github.com/vysogota0399/gophermart_billing/internal/server/entities"
	"github.com/vysogota0399/gophermart_billing/internal/server/services/mocks"
	"github.com/vysogota0399/gophermart_protos/gen/commands/create_order"
	"github.com/vysogota0399/gophermart_protos/gen/common"
	"github.com/vysogota0399/gophermart_protos/gen/entities"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

func TestCreateOrderService_Call(t *testing.T) {
	type fields struct {
		fsm *mocks.MockOrderStateMachineContainer
	}
	type args struct {
		Order *create_order.CreateNewOrderParams
	}
	type want struct {
		err   bool
		order bool
	}
	tests := []struct {
		name    string
		fields  fields
		prepare func(f *fields, in *create_order.CreateNewOrderParams)
		args    args
		want    want
	}{
		{
			name: "when update order state",
			args: args{
				Order: &create_order.CreateNewOrderParams{UploadedAt: timestamppb.Now(), Uuid: &common.Uuid{}, Account: &entities.Account{}},
			},
			prepare: func(f *fields, in *create_order.CreateNewOrderParams) {
				f.fsm.EXPECT().Create(
					gomock.Any(),
					sm.OrderFsmOption{
						Order: &models.Order{
							UUID:       in.Uuid.Value,
							Number:     in.Number,
							UploadedAt: in.UploadedAt.AsTime(),
							AccountID:  in.Account.Id,
						},
					},
				).Return(&models.Order{}, nil)
			},
			want: want{
				order: true,
				err:   false,
			},
		},
		{
			name: "when create order failed",
			args: args{
				Order: &create_order.CreateNewOrderParams{UploadedAt: timestamppb.Now(), Uuid: &common.Uuid{}, Account: &entities.Account{}},
			},
			prepare: func(f *fields, in *create_order.CreateNewOrderParams) {
				f.fsm.EXPECT().Create(
					gomock.Any(),
					sm.OrderFsmOption{
						Order: &models.Order{
							UUID:       in.Uuid.Value,
							Number:     in.Number,
							UploadedAt: in.UploadedAt.AsTime(),
							AccountID:  in.Account.Id,
						},
					},
				).Return(nil, errors.New("error"))
			},
			want: want{
				order: false,
				err:   true,
			},
		},
	}

	lg, err := logging.NewZapLogger(&config.Config{})
	assert.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			fsm := mocks.NewMockOrderStateMachineContainer(ctrl)
			fields := tt.fields
			fields.fsm = fsm
			tt.prepare(&fields, tt.args.Order)

			srv := NewCreateOrderService(lg, fsm)
			order, err := srv.Call(context.Background(), tt.args.Order)

			assert.Equal(t, tt.want.err, err != nil)
			assert.Equal(t, tt.want.order, order != nil)
		})
	}
}
