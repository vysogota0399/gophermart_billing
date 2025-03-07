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
	"github.com/vysogota0399/gophermart_billing/internal/server/commands/mocks"
	"github.com/vysogota0399/gophermart_protos/gen/commands/withdraw"
)

func TestWithdrawCommand_Withdraw(t *testing.T) {
	type fields struct {
		srv *mocks.MockWithdrawService
	}
	type args struct {
		order *models.Order
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
			args:    args{&models.Order{}},
			prepare: func(f *fields, err error) { f.srv.EXPECT().Call(gomock.Any(), gomock.All()).Return(err) },
		},
		{
			name:    "when services failed",
			args:    args{&models.Order{}},
			prepare: func(f *fields, err error) { f.srv.EXPECT().Call(gomock.Any(), gomock.All()).Return(err) },
			wantErr: errors.New("service failed error"),
		},
	}

	lg, err := logging.NewZapLogger(&config.Config{})
	assert.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSrv := mocks.NewMockWithdrawService(ctrl)
			fields := tt.fields
			fields.srv = mockSrv
			tt.prepare(&fields, tt.wantErr)

			comm := NewCreateWithdrawCommand(mockSrv, lg)
			_, err := comm.DoWithdraw(context.Background(), &withdraw.DoWithdrawParams{})
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}
