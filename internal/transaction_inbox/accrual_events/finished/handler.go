package finished

import (
	"context"
	"database/sql"

	"github.com/vysogota0399/gophermart_billing/internal/logging"
	"github.com/vysogota0399/gophermart_billing/internal/models"
	"github.com/vysogota0399/gophermart_billing/internal/transaction_inbox"
	"github.com/vysogota0399/gophermart_protos/gen/events"
	"go.uber.org/zap"
)

type Handler struct {
	lg      *logging.ZapLogger
	command Command
	events  AccrualFinisedEventsRepository
}

type Command interface {
	Call(ctx context.Context, acc *events.FinishedEvent) (*models.Order, error)
}

func NewHandler(
	command Command,
	lg *logging.ZapLogger,
	cfg *transaction_inbox.Config,
	events AccrualFinisedEventsRepository,
) *Handler {
	return &Handler{
		lg:      lg,
		command: command,
		events:  events,
	}
}

type AccrualFinisedEventsRepository interface {
	ReserveEvent(ctx context.Context, eventName string) (*models.AccrualEvent, error)
	SetState(ctx context.Context, uuid string, newState string, tx ...*sql.Tx) error
}

func (h *Handler) Call(ctx context.Context, event *events.FinishedEvent) {
	ctx = h.lg.WithContextFields(ctx, zap.String("actor", "accrual_finised_handler"))

	if err := h.events.SetState(ctx, event.EventUuid, models.AccrualProcessingState); err != nil {
		h.lg.ErrorCtx(ctx, "set event processing state error", zap.Error(err))
		return
	}

	if _, err := h.command.Call(ctx, event); err != nil {
		if err := h.events.SetState(ctx, event.EventUuid, models.AccrualFailedState); err != nil {
			h.lg.ErrorCtx(ctx, "set event failed state error", zap.Error(err))
		}

		h.lg.ErrorCtx(ctx, "command failed with error", zap.Error(err))
		return
	}

	if err := h.events.SetState(ctx, event.EventUuid, models.AccrualFinishedState); err != nil {
		h.lg.ErrorCtx(ctx, "set event finished state error", zap.Error(err))
	}
}
