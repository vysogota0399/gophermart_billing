package server

import (
	"context"
	"net"

	"github.com/vysogota0399/gophermart_billing/internal/config"
	"github.com/vysogota0399/gophermart_protos/gen/commands/withdraw"
	"go.uber.org/fx"
	"google.golang.org/grpc"
)

type AccountingServer struct {
	command withdraw.WithdrawServiceServer
	cfg     *config.Config
	srv     *grpc.Server
}

func (s *AccountingServer) Start() error {
	lis, err := net.Listen("tcp", s.cfg.AccountingServerHTTPAddress)
	if err != nil {
		return err
	}

	withdraw.RegisterWithdrawServiceServer(s.srv, s.command)
	go s.srv.Serve(lis)

	return nil
}

func (s *AccountingServer) Stop() {
	s.srv.GracefulStop()
}

func NewAccountingServer(h withdraw.WithdrawServiceServer, lc fx.Lifecycle, cfg *config.Config) *AccountingServer {
	srv := &AccountingServer{cfg: cfg, srv: grpc.NewServer(), command: h}

	lc.Append(
		fx.Hook{
			OnStart: func(ctx context.Context) error {
				return srv.Start()
			},
			OnStop: func(ctx context.Context) error {
				srv.Stop()
				return nil
			},
		},
	)

	return srv
}
