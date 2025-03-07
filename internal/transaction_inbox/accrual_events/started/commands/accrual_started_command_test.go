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
	"github.com/vysogota0399/gophermart_billing/internal/transaction_inbox/accrual_events/started/commands/mocks"
	"github.com/vysogota0399/gophermart_protos/gen/common"
	events "github.com/vysogota0399/gophermart_protos/gen/events"
)

func TestCreateDebitCommand_Call(t *testing.T) {
	type fields struct {
		fsm *mocks.MockOrderFSM
	}
	type args struct {
		event *events.AccrualStartedEvent
	}
	type want struct {
		err   bool
		order bool
	}
	tests := []struct {
		name    string
		fields  fields
		prepare func(f *fields, in *events.AccrualStartedEvent)
		args    args
		want    want
	}{
		{
			name: "when order state updated",
			args: args{
				event: &events.AccrualStartedEvent{
					EventUuid: &common.Uuid{Value: "event_uuid"},
					OrderUuid: &common.Uuid{Value: "order_uuid"},
				},
			},
			prepare: func(f *fields, in *events.AccrualStartedEvent) {
				f.fsm.EXPECT().Start(
					gomock.Any(),
					entities.OrderFsmOption{
						UUID: in.OrderUuid.Value,
					},
				).Return(&models.Order{}, nil)
			},
			want: want{
				order: true,
				err:   false,
			},
		},
		{
			name: "when update order state failed",
			args: args{
				event: &events.AccrualStartedEvent{
					EventUuid: &common.Uuid{Value: "event_uuid"},
					OrderUuid: &common.Uuid{Value: "order_uuid"},
				},
			},
			prepare: func(f *fields, in *events.AccrualStartedEvent) {
				f.fsm.EXPECT().Start(
					gomock.Any(),
					entities.OrderFsmOption{
						UUID: in.OrderUuid.Value,
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

			fields := tt.fields
			fields.fsm = mocks.NewMockOrderFSM(ctrl)

			tt.prepare(&fields, tt.args.event)

			srv := NewAccrualStartedCommand(fields.fsm, lg)
			order, err := srv.Call(
				context.Background(),
				&events.AccrualStartedEvent{
					EventUuid: tt.args.event.EventUuid,
					OrderUuid: tt.args.event.OrderUuid},
			)
			assert.Equal(t, tt.want.err, err != nil)
			assert.Equal(t, tt.want.order, order != nil)
		})
	}
}
