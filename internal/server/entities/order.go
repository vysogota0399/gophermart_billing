package entities

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/looplab/fsm"
	"github.com/vysogota0399/gophermart_billing/internal/logging"
	"github.com/vysogota0399/gophermart_billing/internal/models"
	"go.uber.org/zap"
)

type OrderFSM struct {
	fsm *fsm.FSM
	rep OrdersRepository
	lg  *logging.ZapLogger
}

type OrdersRepository interface {
	CreateOrder(context.Context, *models.Order, *sql.Tx) error
	UpdateOrderState(ctx context.Context, in *models.Order, tx *sql.Tx) error
	FindByUUIDForUpdate(ctx context.Context, in *models.Order) (*sql.Tx, error)
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

const (
	NewState        = "REGISTERED"
	InvalidState    = "INVALID"
	ProcessingState = "PROCESSING"
	ProcessedState  = "PROCESSED"
)

var OrderStateValues = map[string]int32{
	NewState:        0,
	InvalidState:    1,
	ProcessingState: 2,
	ProcessedState:  3,
}

var OrderStateName = map[int32]string{
	0: NewState,
	1: InvalidState,
	2: ProcessingState,
	3: ProcessedState,
}

var StartEvent = "start"
var CostAccrualeFailedEvent = "cost_cccruale_failed"
var CostAccrualedEvent = "cost_accrualed"

func NewOrderFSM(rep OrdersRepository, lg *logging.ZapLogger) *OrderFSM {
	stateMachine := fsm.NewFSM(
		NewState,
		fsm.Events{
			{Name: StartEvent, Src: []string{NewState}, Dst: ProcessingState},
			{Name: CostAccrualeFailedEvent, Src: []string{ProcessingState, NewState}, Dst: InvalidState},
			{Name: CostAccrualedEvent, Src: []string{ProcessingState}, Dst: ProcessedState},
		},
		fsm.Callbacks{},
	)

	return &OrderFSM{
		fsm: stateMachine,
		lg:  lg,
		rep: rep,
	}
}

func (processor *OrderFSM) new(ctx context.Context, opt OrderFsmOption) (*Order, error) {
	if opt.UUID != "" {
		order := &models.Order{UUID: opt.UUID}
		tx, err := processor.rep.FindByUUIDForUpdate(ctx, order)
		if err != nil {
			return nil, fmt.Errorf("internal/server/entities/order order not found error %w", err)
		}

		return &Order{fsm: processor, Order: order, tx: tx}, nil
	} else if opt.Order != nil {
		tx, err := processor.rep.BeginTx(ctx, &sql.TxOptions{})
		if err != nil {
			return nil, fmt.Errorf("internal/server/entities/order begin tx error %w", err)
		}

		return &Order{fsm: processor, Order: opt.Order, tx: tx}, nil
	}

	return nil, errors.New("internal/server/entities/order invalid params")
}

func (processor *OrderFSM) terminate(ctx context.Context, container *Order) error {
	container.State = OrderStateValues[InvalidState]
	processor.lg.DebugCtx(
		ctx,
		"terminate order",
		zap.Any("order", container),
	)

	return processor.rep.UpdateOrderState(ctx, container.Order, container.tx)
}

func (processor *OrderFSM) send(ctx context.Context, event string, container *Order) error {
	processor.fsm.SetState(OrderStateName[container.Order.State])

	processor.lg.DebugCtx(
		ctx,
		"processing event",
		zap.String("event", event),
		zap.Any("order_before", container),
	)

	if err := processor.fsm.Event(ctx, event); err != nil {
		if err := processor.terminate(ctx, container); err != nil {
			return err
		}

		return fmt.Errorf("internal/server/entities/order invalid state error %w", err)
	}

	container.State = OrderStateValues[processor.fsm.Current()]

	processor.lg.DebugCtx(
		ctx,
		"processing event - new order state",
		zap.String("event", event),
		zap.Any("order_after", container),
	)

	return nil
}

func (processor *OrderFSM) CostAccrualed(ctx context.Context, amount int64, opt OrderFsmOption) (*models.Order, error) {
	container, err := processor.new(ctx, opt)
	if err != nil {
		return nil, err
	}
	defer container.rollback()

	if err := processor.send(ctx, CostAccrualedEvent, container); err != nil {
		return nil, err
	}

	if err := processor.rep.UpdateOrderState(ctx, container.Order, container.tx); err != nil {
		if err := processor.terminate(ctx, container); err != nil {
			return nil, err
		}

		return nil, err
	}

	container.commit()
	return container.Order, nil
}

func (processor *OrderFSM) CostAccrualFailed(ctx context.Context, opt OrderFsmOption) (*models.Order, error) {
	container, err := processor.new(ctx, opt)
	if err != nil {
		return nil, err
	}
	defer container.rollback()

	if err := processor.send(ctx, CostAccrualeFailedEvent, container); err != nil {
		return nil, err
	}

	if err := processor.rep.UpdateOrderState(ctx, container.Order, container.tx); err != nil {
		return nil, fmt.Errorf("internal/server/entities/order update order state error %w", err)
	}

	container.commit()
	return container.Order, nil
}

func (processor *OrderFSM) Create(ctx context.Context, opt OrderFsmOption) (*models.Order, error) {
	container, err := processor.new(ctx, opt)
	defer container.rollback()
	if err != nil {
		return nil, err
	}

	container.State = OrderStateValues[NewState]

	if err := processor.rep.CreateOrder(ctx, container.Order, container.tx); err != nil {
		return nil, fmt.Errorf("internal/server/entities/order create order error %w", err)
	}

	container.commit()
	return container.Order, nil
}

func (processor *OrderFSM) Start(ctx context.Context, opt OrderFsmOption) (*models.Order, error) {
	container, err := processor.new(ctx, opt)
	if err != nil {
		return nil, err
	}
	defer container.rollback()

	if err := processor.send(ctx, StartEvent, container); err != nil {
		return nil, err
	}

	if err := container.fsm.rep.UpdateOrderState(ctx, container.Order, container.tx); err != nil {
		return nil, fmt.Errorf("internal/server/entities/order update order state error %w", err)
	}

	container.commit()
	return container.Order, nil
}

type Order struct {
	*models.Order
	fsm *OrderFSM
	tx  *sql.Tx
}

type OrderFsmOption struct {
	Order *models.Order
	UUID  string
}

func (container *Order) commit() {
	container.tx.Commit()
}

func (container *Order) rollback() {
	container.tx.Rollback()
}
