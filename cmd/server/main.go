package main

import (
	main_config "github.com/vysogota0399/gophermart_billing/internal/config"
	"github.com/vysogota0399/gophermart_billing/internal/logging"
	"github.com/vysogota0399/gophermart_billing/internal/repositories"
	"github.com/vysogota0399/gophermart_billing/internal/server"
	"github.com/vysogota0399/gophermart_billing/internal/server/commands"
	"github.com/vysogota0399/gophermart_billing/internal/server/entities"
	"github.com/vysogota0399/gophermart_billing/internal/server/services"
	"github.com/vysogota0399/gophermart_billing/internal/storage"
	"github.com/vysogota0399/gophermart_billing/internal/transaction_inbox"
	"github.com/vysogota0399/gophermart_billing/internal/transaction_outbox"
	order_event_creaded "github.com/vysogota0399/gophermart_billing/internal/transaction_outbox/order_events/created"
	order_event_updated "github.com/vysogota0399/gophermart_billing/internal/transaction_outbox/order_events/updated"
	"github.com/vysogota0399/gophermart_billing/internal/transaction_outbox/transaction_events/processed"
	"github.com/vysogota0399/gophermart_protos/gen/commands/create_order"
	"github.com/vysogota0399/gophermart_protos/gen/commands/withdraw"
	"go.uber.org/fx"
)

func main() {
	fx.New(CreateApp()).Run()
}

func CreateApp() fx.Option {
	return fx.Options(
		fx.Provide(
			logging.NewZapLogger,
			logging.NewKafkaErrorLogger,
			logging.NewKafkaLogger,

			storage.NewStorage,
			server.NewOrdersServer,
			server.NewAccountingServer,

			fx.Annotate(entities.NewOrderFSM, fx.As(new(services.OrderStateMachineContainer))),
			fx.Annotate(repositories.NewOutboxEventsRepository, fx.As(new(repositories.OrderEvents))),
			fx.Annotate(repositories.NewOrdersRepository, fx.As(new(entities.OrdersRepository))),
			fx.Annotate(services.NewCreateOrderService, fx.As(new(commands.CreateOrderService))),
			fx.Annotate(commands.NewCreateOrderCommand, fx.As(new(create_order.CreateOrderServiceServer))),

			// сервер
			fx.Annotate(repositories.NewOutboxEventsRepository, fx.As(new(repositories.TransactionEvents))),
			fx.Annotate(repositories.NewAccountingRepository, fx.As(new(services.WithdrawRepository))),
			fx.Annotate(services.NewWithdrawService, fx.As(new(commands.WithdrawService))),
			fx.Annotate(commands.NewCreateWithdrawCommand, fx.As(new(withdraw.WithdrawServiceServer))),

			// (transaction_outbox) паблишер событий создания заказов
			order_event_creaded.NewDaemon,
			fx.Annotate(order_event_creaded.NewPublisher, fx.As(new(order_event_creaded.OrderCreatedEventsPublisher))),
			fx.Annotate(repositories.NewOutboxEventsRepository, fx.As(new(order_event_creaded.OrderCreatedEventsRepository))),

			// (transaction_outbox) паблишер событий обновления заказов
			order_event_updated.NewDaemon,
			fx.Annotate(order_event_updated.NewPublisher, fx.As(new(order_event_updated.OrderUpdatedEventsPublisher))),
			fx.Annotate(repositories.NewOutboxEventsRepository, fx.As(new(order_event_updated.OrderUpdatedEventsRepository))),

			// (transaction_outbox) паблишер событий создания транзакций
			processed.NewDaemon,
			fx.Annotate(processed.NewPublisher, fx.As(new(processed.TransactionProcessedPublisher))),
			fx.Annotate(repositories.NewOutboxEventsRepository, fx.As(new(processed.TransactionProcessedEventsRepository))),
		),
		fx.Supply(
			main_config.MustNewConfig(),
			transaction_outbox.MustNewConfig(),
			transaction_inbox.MustNewConfig(),
		),
		fx.Invoke(
			checkDBConnection,
			startOrdersServer,
			startAccountingServer,
			startTransactionOutboxAccountingCreated,
			startTransactionOutboxOrderCreated,
			startTransactionOutboxOrderUpdated,
		),
	)
}

func startOrdersServer(*server.OrdersServer)                         {}
func checkDBConnection(*storage.Storage)                             {}
func startAccountingServer(*server.AccountingServer)                 {}
func startTransactionOutboxOrderCreated(*order_event_creaded.Daemon) {}
func startTransactionOutboxOrderUpdated(*order_event_updated.Daemon) {}
func startTransactionOutboxAccountingCreated(*processed.Daemon)      {}
