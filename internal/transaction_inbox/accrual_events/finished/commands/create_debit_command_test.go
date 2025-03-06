package commands

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/vysogota0399/gophermart_billing/internal/config"
	"github.com/vysogota0399/gophermart_billing/internal/logging"
	"github.com/vysogota0399/gophermart_billing/internal/models"
	"github.com/vysogota0399/gophermart_billing/internal/server/entities"
	"github.com/vysogota0399/gophermart_billing/internal/transaction_inbox/accrual_events/finished/commands/mocks"
	"github.com/vysogota0399/gophermart_protos/gen/common"
	events "github.com/vysogota0399/gophermart_protos/gen/events"
	"google.golang.org/genproto/googleapis/type/money"
)

func TestCreateDebitCommand_Call(t *testing.T) {
	type fields struct {
		fsm        *mocks.MockAccrualCreator
		accounting *mocks.MockDebitCreator
	}
	type args struct {
		event *events.AccrualFinishedEvent
	}
	type want struct {
		err   bool
		order bool
	}
	tests := []struct {
		name    string
		fields  fields
		prepare func(f *fields, in *events.AccrualFinishedEvent)
		args    args
		want    want
	}{
		{
			name: "when order state updated",
			args: args{
				event: &events.AccrualFinishedEvent{
					EventUuid: &common.Uuid{Value: "event_uuid"},
					OrderUuid: &common.Uuid{Value: "order_uuid"},
				},
			},
			prepare: func(f *fields, in *events.AccrualFinishedEvent) {
				f.fsm.EXPECT().CostAccrualed(
					gomock.Any(),
					gomock.Any(),
					entities.OrderFsmOption{
						UUID: in.OrderUuid.Value,
					},
				).Return(&models.Order{}, nil)
				f.accounting.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
			},
			want: want{
				order: true,
				err:   false,
			},
		},
		{
			name: "when order state updated failed",
			args: args{
				event: &events.AccrualFinishedEvent{
					EventUuid: &common.Uuid{Value: "event_uuid"},
					OrderUuid: &common.Uuid{Value: "order_uuid"},
				},
			},
			prepare: func(f *fields, in *events.AccrualFinishedEvent) {
				f.fsm.EXPECT().CostAccrualed(
					gomock.Any(),
					gomock.Any(),
					entities.OrderFsmOption{
						UUID: in.OrderUuid.Value,
					},
				).Return(nil, errors.New("error"))
				f.accounting.EXPECT().Create(gomock.Any(), gomock.Any()).Times(0)
			},
			want: want{
				order: false,
				err:   true,
			},
		},
		{
			name: "when accounting failed",
			args: args{
				event: &events.AccrualFinishedEvent{
					EventUuid: &common.Uuid{Value: "event_uuid"},
					OrderUuid: &common.Uuid{Value: "order_uuid"},
				},
			},
			prepare: func(f *fields, in *events.AccrualFinishedEvent) {
				f.fsm.EXPECT().CostAccrualed(
					gomock.Any(),
					gomock.Any(),
					entities.OrderFsmOption{
						UUID: in.OrderUuid.Value,
					},
				).Return(&models.Order{}, nil)

				f.accounting.EXPECT().Create(gomock.Any(), gomock.Any()).Return(errors.New("error"))
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

			fields := tt.fields
			fields.accounting = mocks.NewMockDebitCreator(ctrl)
			fields.fsm = mocks.NewMockAccrualCreator(ctrl)

			tt.prepare(&fields, tt.args.event)

			srv := NewCreateDebitCommand(fields.fsm, fields.accounting, lg)
			order, err := srv.Call(
				context.Background(),
				&events.AccrualFinishedEvent{
					EventUuid: tt.args.event.EventUuid,
					OrderUuid: tt.args.event.OrderUuid,
					Amount:    &money.Money{Units: 1},
				},
			)
			assert.Equal(t, tt.want.err, err != nil)
			assert.Equal(t, tt.want.order, order != nil)
		})
	}
}
