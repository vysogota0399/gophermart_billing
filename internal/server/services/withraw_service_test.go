package services

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/vysogota0399/gophermart_billing/internal/config"
	"github.com/vysogota0399/gophermart_billing/internal/logging"
	"github.com/vysogota0399/gophermart_billing/internal/server/services/mocks"
	"github.com/vysogota0399/gophermart_protos/gen/commands/withdraw"
	"google.golang.org/genproto/googleapis/type/money"
)

func TestWithdrawService_Call(t *testing.T) {
	type fields struct {
		wdRep *mocks.MockWithdrawRepository
	}
	type args struct {
		wd *withdraw.WithdrawParams
	}
	type want struct {
		err error
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    want
		prepare func(*fields, error, *WithdrawService)
	}{
		{
			name: "when calc balance error",
			prepare: func(f *fields, err error, srv *WithdrawService) {
				var balance int64
				tx := &sql.Tx{}
				f.wdRep.EXPECT().BeginTX(gomock.Any(), gomock.Any()).Return(tx, nil)
				f.wdRep.EXPECT().BalanceForUpdate(gomock.Any(), gomock.Any(), tx).Return(balance, err)
				f.wdRep.EXPECT().RollbackTX(tx)
			},
			want: want{
				err: errors.New("calc balance error"),
			},
			args: args{&withdraw.WithdrawParams{}},
		},
		{
			name: "when invalid balance error",
			prepare: func(f *fields, err error, srv *WithdrawService) {
				var balance int64
				tx := &sql.Tx{}
				f.wdRep.EXPECT().BeginTX(gomock.Any(), gomock.Any()).Return(tx, nil)
				f.wdRep.EXPECT().BalanceForUpdate(gomock.Any(), gomock.Any(), tx).Return(balance, nil)
				f.wdRep.EXPECT().RollbackTX(tx)
			},
			want: want{
				err: ErrNotEnoughFunds,
			},
			args: args{&withdraw.WithdrawParams{Amount: &money.Money{Units: 9999}}},
		},
		{
			name: "when save withdraw error",
			prepare: func(f *fields, err error, srv *WithdrawService) {
				var balance int64 = 10
				tx := &sql.Tx{}
				f.wdRep.EXPECT().BeginTX(gomock.Any(), gomock.Any()).MaxTimes(int(srv.retryCount)).Return(tx, nil)
				f.wdRep.EXPECT().BalanceForUpdate(gomock.Any(), gomock.Any(), tx).MaxTimes(int(srv.retryCount)).Return(balance, nil)
				f.wdRep.EXPECT().CreateTX(gomock.Any(), gomock.Any(), tx).MaxTimes(int(srv.retryCount)).Return(err)
				f.wdRep.EXPECT().RollbackTX(tx).MaxTimes(int(srv.retryCount))
			},
			want: want{
				err: &pgconn.PgError{Code: pgerrcode.SerializationFailure},
			},
			args: args{&withdraw.WithdrawParams{Amount: &money.Money{Units: 0}}},
		},
		{
			name: "when succeeded",
			prepare: func(f *fields, err error, srv *WithdrawService) {
				var balance int64 = 10
				tx := &sql.Tx{}
				f.wdRep.EXPECT().BeginTX(gomock.Any(), gomock.Any()).Return(tx, nil)
				f.wdRep.EXPECT().BalanceForUpdate(gomock.Any(), gomock.Any(), tx).Return(balance, nil)
				f.wdRep.EXPECT().CreateTX(gomock.Any(), gomock.Any(), tx).Return(nil)
				f.wdRep.EXPECT().RollbackTX(tx).Times(1)
				f.wdRep.EXPECT().CommitTX(tx).Times(1)
			},
			want: want{},
			args: args{&withdraw.WithdrawParams{Amount: &money.Money{Units: 0}}},
		},
	}

	lg, err := logging.NewZapLogger(&config.Config{})
	assert.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			fields := tt.fields
			fields.wdRep = mocks.NewMockWithdrawRepository(ctrl)
			srv := NewWithdrawService(fields.wdRep, lg)

			tt.prepare(&fields, tt.want.err, srv)

			err := srv.Call(context.Background(), tt.args.wd)

			assert.ErrorIs(t, err, tt.want.err)
		})
	}
}
