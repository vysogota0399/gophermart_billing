package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/vysogota0399/gophermart_billing/internal/config"
	"github.com/vysogota0399/gophermart_billing/internal/logging"
	"github.com/vysogota0399/gophermart_billing/internal/models"
	"github.com/vysogota0399/gophermart_billing/internal/server/entities"
	"github.com/vysogota0399/gophermart_billing/internal/server/services/mocks"
	"github.com/vysogota0399/gophermart_protos/gen/commands/create_order"
)

func TestCreateOrderService_Call(t *testing.T) {
	type fields struct {
		fsm *mocks.MockOrderStateMachineContainer
	}
	type args struct {
		Order *create_order.NewOrder
	}
	type want struct {
		err   bool
		order bool
	}
	tests := []struct {
		name    string
		fields  fields
		prepare func(f *fields, in *create_order.NewOrder)
		args    args
		want    want
	}{
		{
			name: "when update order state",
			args: args{
				Order: &create_order.NewOrder{UploadedAt: time.Now().Format(time.RFC3339Nano)},
			},
			prepare: func(f *fields, in *create_order.NewOrder) {
				t, _ := time.Parse(time.RFC3339Nano, in.UploadedAt)
				f.fsm.EXPECT().Create(
					gomock.Any(),
					entities.OrderFsmOption{
						Order: &models.Order{
							UUID:       in.Uuid,
							Number:     in.Number,
							UploadedAt: t,
							AccountID:  in.AccountId,
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
				Order: &create_order.NewOrder{UploadedAt: time.Now().Format(time.RFC3339Nano)},
			},
			prepare: func(f *fields, in *create_order.NewOrder) {
				t, _ := time.Parse(time.RFC3339Nano, in.UploadedAt)
				f.fsm.EXPECT().Create(
					gomock.Any(),
					entities.OrderFsmOption{
						Order: &models.Order{
							UUID:       in.Uuid,
							Number:     in.Number,
							UploadedAt: t,
							AccountID:  in.AccountId,
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
