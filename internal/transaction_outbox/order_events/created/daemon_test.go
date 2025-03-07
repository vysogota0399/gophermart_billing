package created

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/vysogota0399/gophermart_billing/internal/config"
	"github.com/vysogota0399/gophermart_billing/internal/logging"
	"github.com/vysogota0399/gophermart_billing/internal/models"
	"github.com/vysogota0399/gophermart_billing/internal/transaction_outbox/order_events/created/mocks"
)

func TestDaemon_processEvent(t *testing.T) {
	type fields struct {
		events *mocks.MockOrderCreatedEventsRepository
		pub    *mocks.MockOrderCreatedEventsPublisher
	}
	tests := []struct {
		name      string
		prepare   func(f *fields)
		fields    fields
		wantError bool
	}{
		{
			name: "when begin tx failed",
			prepare: func(f *fields) {
				f.events.EXPECT().BeginTx(gomock.Any(), gomock.Any()).Return(nil, errors.New("error"))
			},
			wantError: true,
		},
		{
			name: "when search event failed",
			prepare: func(f *fields) {
				f.events.EXPECT().BeginTx(gomock.Any(), gomock.Any()).Return(&sql.Tx{}, nil)
				f.events.EXPECT().ReserveNewOrderEvent(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("error"))
				f.events.EXPECT().RollbackTX(gomock.Any()).Return(nil)
			},
			wantError: true,
		},
		{
			name: "when events not found",
			prepare: func(f *fields) {
				f.events.EXPECT().BeginTx(gomock.Any(), gomock.Any()).Return(&sql.Tx{}, nil)
				f.events.EXPECT().ReserveNewOrderEvent(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil)
				f.events.EXPECT().CommitTX(gomock.Any()).Return(nil)
				f.events.EXPECT().RollbackTX(gomock.Any()).Times(1)
			},
			wantError: false,
		},
		{
			name: "when publisher failed",
			prepare: func(f *fields) {
				event := &models.OrderEvent{}
				f.events.EXPECT().BeginTx(gomock.Any(), gomock.Any()).Return(&sql.Tx{}, nil)
				f.events.EXPECT().ReserveNewOrderEvent(gomock.Any(), gomock.Any(), gomock.Any()).Return(event, nil)
				f.pub.EXPECT().Publish(gomock.Any(), event).Return(errors.New("error"))
				f.events.EXPECT().RollbackTX(gomock.Any()).Return(nil)
			},
			wantError: true,
		},
		{
			name: "when update event new state failed",
			prepare: func(f *fields) {
				event := &models.OrderEvent{}
				tx := &sql.Tx{}
				f.events.EXPECT().BeginTx(gomock.Any(), gomock.Any()).Return(tx, nil)
				f.events.EXPECT().ReserveNewOrderEvent(gomock.Any(), gomock.Any(), gomock.Any()).Return(event, nil)
				f.pub.EXPECT().Publish(gomock.Any(), event).Return(nil)
				f.events.EXPECT().Send(gomock.Any(), event, tx).Return(errors.New("error"))
				f.events.EXPECT().RollbackTX(gomock.Any()).Return(nil)
			},
			wantError: true,
		},
		{
			name: "when succeeded",
			prepare: func(f *fields) {
				event := &models.OrderEvent{}
				tx := &sql.Tx{}
				f.events.EXPECT().BeginTx(gomock.Any(), gomock.Any()).Return(tx, nil)
				f.events.EXPECT().ReserveNewOrderEvent(gomock.Any(), gomock.Any(), gomock.Any()).Return(event, nil)
				f.pub.EXPECT().Publish(gomock.Any(), event).Return(nil)
				f.events.EXPECT().Send(gomock.Any(), event, tx).Return(nil)
				f.events.EXPECT().CommitTX(gomock.Any()).Return(nil)
				f.events.EXPECT().RollbackTX(gomock.Any()).Times(1)
			},
			wantError: false,
		},
	}

	lg, err := logging.NewZapLogger(&config.Config{})
	assert.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			fields := tt.fields
			rep := mocks.NewMockOrderCreatedEventsRepository(ctrl)
			pub := mocks.NewMockOrderCreatedEventsPublisher(ctrl)

			fields.events = rep
			fields.pub = pub

			tt.prepare(&fields)

			dmn := &Daemon{
				lg:        lg,
				events:    rep,
				publisher: pub,
			}

			err := dmn.processEvent(context.Background())
			assert.Equal(t, tt.wantError, err != nil)
		})
	}
}
