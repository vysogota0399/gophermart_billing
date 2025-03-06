package commands

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/vysogota0399/gophermart_billing/internal/config"
	"github.com/vysogota0399/gophermart_billing/internal/logging"
	"github.com/vysogota0399/gophermart_billing/internal/server/commands/mocks"
	"github.com/vysogota0399/gophermart_protos/gen/commands/create_order"
)

func TestCreateOrderCommand_CreateOrder(t *testing.T) {
	var ErrStub = errors.New("service error")
	type fields struct {
		srv *mocks.MockCreateOrderService
	}
	type args struct {
		order *create_order.CreateNewOrderParams
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
		prepare func(f *fields, err error)
	}{
		{
			name:    "when services succeeded",
			args:    args{&create_order.CreateNewOrderParams{}},
			prepare: func(f *fields, err error) { f.srv.EXPECT().Call(gomock.Any(), gomock.All()).Return(nil, err) },
		},
		{
			name:    "when services failed",
			args:    args{&create_order.CreateNewOrderParams{}},
			prepare: func(f *fields, err error) { f.srv.EXPECT().Call(gomock.Any(), gomock.All()).Return(nil, err) },
			wantErr: ErrStub,
		},
		{
			name:    "when order with same number already exists",
			args:    args{&create_order.CreateNewOrderParams{}},
			prepare: func(f *fields, err error) { f.srv.EXPECT().Call(gomock.Any(), gomock.All()).Return(nil, err) },
			wantErr: ErrOrderAlreadtExists,
		},
	}

	lg, err := logging.NewZapLogger(&config.Config{})
	assert.NoError(t, err)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSrv := mocks.NewMockCreateOrderService(ctrl)
			fields := tt.fields
			fields.srv = mockSrv
			tt.prepare(&fields, tt.wantErr)

			comm := NewCreateOrderCommand(mockSrv, lg)
			_, err := comm.Create(context.Background(), &create_order.CreateNewOrderParams{})
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}
