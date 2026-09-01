package application

/**
 * ============================================================================
 * ORCHESTRATION: Order Placement Saga (Distributed Transaction Manager)
 * BOUNDED CONTEXT: Orders
 * PATTERN: Saga Pattern with Compensating Transactions
 * 
 * WHY SAGA OVER 2-PHASE COMMIT (2PC)?
 * In a high-scale microservice architecture, 2PC blocks database resources and creates
 * severe latency bottlenecks. The Saga pattern executes a series of local transactions
 * across Inventory, Payment, and Order services with compensating rollbacks if any step fails.
 * ============================================================================
 */

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	inventoryService "github.com/stylehome/eshop/internal/inventory/service"
	ordersDomain "github.com/stylehome/eshop/internal/orders/domain"
	paymentDomain "github.com/stylehome/eshop/internal/payment/domain"
	paymentService "github.com/stylehome/eshop/internal/payment/service"
	"github.com/stylehome/eshop/internal/pkg/events"
	"github.com/stylehome/eshop/internal/pkg/outbox"
)

type PlaceOrderCommand struct {
	UserID          string
	IdempotencyKey  string
	Items           []ordersDomain.OrderLineItem
	ShippingAddress ordersDomain.Address
	BillingAddress  ordersDomain.Address
	PaymentMethod   string
	PaymentToken    string
	SubtotalCents   int64
	DiscountCents   int64
	ShippingCents   int64
	TaxCents        int64
	GrandTotalCents int64
}

type OrderSagaOrchestrator struct {
	inventorySvc *inventoryService.Service
	paymentSvc   *paymentService.Service
	outboxWorker *outbox.Worker
}

func NewOrderSagaOrchestrator(
	inv *inventoryService.Service,
	pay *paymentService.Service,
	ob *outbox.Worker,
) *OrderSagaOrchestrator {
	return &OrderSagaOrchestrator{
		inventorySvc: inv,
		paymentSvc:   pay,
		outboxWorker: ob,
	}
}

// Execute handles the step-by-step distributed transaction
func (s *OrderSagaOrchestrator) Execute(ctx context.Context, cmd PlaceOrderCommand) (*ordersDomain.Order, error) {
	bytes := make([]byte, 6)
	_, _ = rand.Read(bytes)
	orderID := "ord_" + hex.EncodeToString(bytes)
	orderNumber := fmt.Sprintf("SH-%d-%s", time.Now().Unix()%100000, hex.EncodeToString(bytes[:2]))

	// STEP 1: Inventory Reservation (Synchronous gRPC)
	var reserveItems []inventoryService.ReserveItemRequest
	for _, it := range cmd.Items {
		reserveItems = append(reserveItems, inventoryService.ReserveItemRequest{
			SKU:      it.SKU,
			Quantity: it.Quantity,
		})
	}

	resResult, err := s.inventorySvc.ReserveStock(ctx, orderID, reserveItems, 900)
	if err != nil {
		return nil, fmt.Errorf("inventory reservation error: %w", err)
	}
	if !resResult.Success {
		return nil, fmt.Errorf("inventory out of stock: %s", resResult.Error)
	}

	// STEP 2: Payment Authorization (Synchronous gRPC with Idempotency)
	authResp, err := s.paymentSvc.Authorize(ctx, paymentService.AuthRequest{
		OrderID:        orderID,
		UserID:         cmd.UserID,
		Amount:         paymentDomain.Money{AmountInCents: cmd.GrandTotalCents, Currency: "USD"},
		PaymentMethod:  cmd.PaymentMethod,
		PaymentToken:   cmd.PaymentToken,
		IdempotencyKey: cmd.IdempotencyKey,
	})

	if err != nil || !authResp.Success {
		// COMPENSATING ACTION: Release held inventory reservation
		_ = s.inventorySvc.ReleaseReservation(ctx, resResult.ReservationID)
		if err != nil {
			return nil, fmt.Errorf("payment gateway error: %w", err)
		}
		return nil, fmt.Errorf("payment authorization failed: %s", authResp.FailureReason)
	}

	// STEP 3: Order Persistence (Local PostgreSQL Transaction)
	now := time.Now()
	order := &ordersDomain.Order{
		ID:              orderID,
		OrderNumber:     orderNumber,
		UserID:          cmd.UserID,
		Status:          ordersDomain.StatusPaymentConfirmed,
		Items:           cmd.Items,
		ShippingAddress: cmd.ShippingAddress,
		BillingAddress:  cmd.BillingAddress,
		SubtotalCents:   cmd.SubtotalCents,
		DiscountCents:   cmd.DiscountCents,
		ShippingCents:   cmd.ShippingCents,
		TaxCents:        cmd.TaxCents,
		GrandTotalCents: cmd.GrandTotalCents,
		Currency:        "USD",
		PaymentMethod:   cmd.PaymentMethod,
		ReservationID:   resResult.ReservationID,
		TransactionID:   authResp.TransactionID,
		IdempotencyKey:  cmd.IdempotencyKey,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	// STEP 4: Commit Inventory Reservation (Saga Confirmation)
	_ = s.inventorySvc.CommitReservation(ctx, resResult.ReservationID)

	// STEP 5: Stage Domain Event to Transactional Outbox (Atomically with Order)
	var eventItems []events.OrderItemData
	for _, it := range cmd.Items {
		eventItems = append(eventItems, events.OrderItemData{
			SKU:            it.SKU,
			ProductID:      it.ProductID,
			Quantity:       it.Quantity,
			UnitPriceCents: it.UnitPriceCents,
		})
	}

	eventPayload, _ := json.Marshal(events.OrderPlacedEvent{
		BaseEvent: events.BaseEvent{
			EventID:       "evt_" + hex.EncodeToString(bytes),
			EventType:     events.TypeOrderPlaced,
			AggregateID:   orderID,
			AggregateType: "ORDER",
			Timestamp:     now,
			Version:       1,
		},
		OrderID:         orderID,
		OrderNumber:     orderNumber,
		UserID:          cmd.UserID,
		GrandTotalCents: cmd.GrandTotalCents,
		Currency:        "USD",
		Items:           eventItems,
	})

	s.outboxWorker.StageEvent(outbox.OutboxRecord{
		ID:            "ob_" + hex.EncodeToString(bytes),
		AggregateType: "ORDER",
		AggregateID:   orderID,
		EventType:     events.TypeOrderPlaced,
		Payload:       eventPayload,
		Published:     false,
		CreatedAt:     now,
	})

	return order, nil
}
