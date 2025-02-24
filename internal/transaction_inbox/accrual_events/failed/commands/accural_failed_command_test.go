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
	"github.com/vysogota0399/gophermart_billing/internal/transaction_inbox/accrual_events/failed/commands/mocks"
	events "github.com/vysogota0399/gophermart_protos/gen/events"
)

func TestAccrualFailedCommand_Call(t *testing.T) {
	type fields struct {
		fsm *mocks.MockAccrualFailder
	}
	type args struct {
		event *events.FailedEvent
	}
	type want struct {
		err   bool
		order bool
	}
	tests := []struct {
		name    string
		fields  fields
		prepare func(f *fields, in *events.FailedEvent)
		args    args
		want    want
	}{

		{
			name: "when order state updated",
			args: args{
				event: &events.FailedEvent{EventUuid: "event_uuid", OrderUuid: "order_uuid"},
			},
			prepare: func(f *fields, in *events.FailedEvent) {
				f.fsm.EXPECT().CostAccrualFailed(
					gomock.Any(),
					entities.OrderFsmOption{
						UUID: in.OrderUuid,
					},
				).Return(&models.Order{}, nil)
			},
			want: want{
				err:   false,
				order: true,
			},
		},
		{
			name: "when order state updated",
			args: args{
				event: &events.FailedEvent{EventUuid: "event_uuid", OrderUuid: "order_uuid"},
			},
			prepare: func(f *fields, in *events.FailedEvent) {
				f.fsm.EXPECT().CostAccrualFailed(
					gomock.Any(),
					entities.OrderFsmOption{
						UUID: in.OrderUuid,
					},
				).Return(nil, errors.New("error"))
			},
			want: want{
				err:   true,
				order: false,
			},
		},
	}

	lg, err := logging.NewZapLogger(&config.Config{})
	assert.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			fsm := mocks.NewMockAccrualFailder(ctrl)
			fields := tt.fields
			fields.fsm = fsm

			tt.prepare(&fields, tt.args.event)

			srv := NewAccrualFailedCommand(fields.fsm, lg)
			order, err := srv.Call(context.Background(), tt.args.event)
			assert.Equal(t, tt.want.err, err != nil)
			assert.Equal(t, tt.want.order, order != nil)
		})
	}
}
