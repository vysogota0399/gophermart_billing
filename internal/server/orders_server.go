package server

import (
	"context"
	"net"

	"github.com/vysogota0399/gophermart_billing/internal/config"
	"github.com/vysogota0399/gophermart_protos/gen/commands/create_order"
	"go.uber.org/fx"
	"google.golang.org/grpc"
)

type OrdersServer struct {
	command create_order.CreateOrderServiceServer
	cfg     *config.Config
	srv     *grpc.Server
}

func (s *OrdersServer) Start() error {
	lis, err := net.Listen("tcp", s.cfg.OrdersServerHTTPAddress)
	if err != nil {
		return err
	}

	create_order.RegisterCreateOrderServiceServer(s.srv, s.command)
	go s.srv.Serve(lis)

	return nil
}

func (s *OrdersServer) Stop() {
	s.srv.GracefulStop()
}

func NewOrdersServer(h create_order.CreateOrderServiceServer, lc fx.Lifecycle, cfg *config.Config) *OrdersServer {
	srv := &OrdersServer{cfg: cfg, srv: grpc.NewServer(), command: h}

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
