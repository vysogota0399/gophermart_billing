package main

import (
	"github.com/vysogota0399/gophermart_billing/internal/config"
	"github.com/vysogota0399/gophermart_billing/internal/logging"
	"github.com/vysogota0399/gophermart_billing/internal/repositories"
	"github.com/vysogota0399/gophermart_billing/internal/server/entities"
	"github.com/vysogota0399/gophermart_billing/internal/storage"
	"github.com/vysogota0399/gophermart_billing/internal/transaction_inbox"
	"github.com/vysogota0399/gophermart_billing/internal/transaction_inbox/accrual_events"
	"github.com/vysogota0399/gophermart_billing/internal/transaction_inbox/accrual_events/failed"
	failed_commands "github.com/vysogota0399/gophermart_billing/internal/transaction_inbox/accrual_events/failed/commands"
	"github.com/vysogota0399/gophermart_billing/internal/transaction_inbox/accrual_events/finished"
	finished_commands "github.com/vysogota0399/gophermart_billing/internal/transaction_inbox/accrual_events/finished/commands"
	"github.com/vysogota0399/gophermart_billing/internal/transaction_inbox/accrual_events/started"
	started_commands "github.com/vysogota0399/gophermart_billing/internal/transaction_inbox/accrual_events/started/commands"
	"go.uber.org/fx"
)

func main() {
	fx.New(CreateApp()).Run()
}

func CreateApp() fx.Option {
	return fx.Options(
		fx.Provide(
			logging.NewZapLogger,
			storage.NewStorage,
			logging.NewKafkaErrorLogger,
			logging.NewKafkaLogger,
			accrual_events.NewConsumer,

			fx.Annotate(repositories.NewInboxEventsRepository, fx.As(new(accrual_events.InboxEventsRepository))),

			fx.Annotate(failed.NewHandler, fx.As(new(accrual_events.AccrualFailedHandler))),
			fx.Annotate(failed_commands.NewAccrualFailedCommand, fx.As(new(failed.Command))),
			fx.Annotate(repositories.NewInboxEventsRepository, fx.As(new(failed.AccrualFailedEventsRepository))),
			fx.Annotate(entities.NewOrderFSM, fx.As(new(failed_commands.AccrualFailder))),
			fx.Annotate(repositories.NewOrdersRepository, fx.As(new(entities.OrdersRepository))),
			fx.Annotate(repositories.NewOutboxEventsRepository, fx.As(new(repositories.TransactionEvents))),

			fx.Annotate(finished.NewHandler, fx.As(new(accrual_events.AccrualFinishedHandler))),
			fx.Annotate(finished_commands.NewCreateDebitCommand, fx.As(new(finished.Command))),
			fx.Annotate(repositories.NewInboxEventsRepository, fx.As(new(finished.AccrualFinisedEventsRepository))),
			fx.Annotate(repositories.NewOutboxEventsRepository, fx.As(new(repositories.OrderEvents))),
			fx.Annotate(repositories.NewAccountingRepository, fx.As(new(finished_commands.DebitCreator))),
			fx.Annotate(entities.NewOrderFSM, fx.As(new(finished_commands.AccrualCreator))),

			fx.Annotate(started.NewHandler, fx.As(new(accrual_events.AccrualStartedHandler))),
			fx.Annotate(started_commands.NewAccrualStartedCommand, fx.As(new(started.Command))),
			fx.Annotate(repositories.NewInboxEventsRepository, fx.As(new(started.AccrualStartedEventsRepository))),
			fx.Annotate(entities.NewOrderFSM, fx.As(new(started_commands.OrderFSM))),
		),
		fx.Supply(
			transaction_inbox.MustNewConfig(),
			config.MustNewConfig(),
		),
		fx.Invoke(
			startConsumer,
		),
	)
}

func startConsumer(*accrual_events.Consumer) {}
