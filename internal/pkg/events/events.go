package events

/**
 * ============================================================================
 * PACKAGE: internal/pkg/events
 * TECHNOLOGY: Apache Kafka Event Payload Definitions
 * PATTERN: Domain Events & Event-Driven Architecture (EDA)
 * 
 * WHY APACHE KAFKA DOMAIN EVENTS?
 * 1. Asynchronous Decoupling: Orders service can confirm order placement to the customer
 *    in <50ms without waiting for slow operations like sending confirmation emails,
 *    updating recommendation engines, or synchronizing accounting ERP systems.
 * 2. Event Streaming Durability: Kafka retains immutable event logs, allowing new
 *    microservices (like Analytics or Fraud Detection) to replay historical events.
 * ============================================================================
 */

import "time"

// Event Names
const (
	TopicOrderEvents     = "eshop.order.events"
	TopicPaymentEvents   = "eshop.payment.events"
	TopicInventoryEvents = "eshop.inventory.events"
	TopicCatalogEvents   = "eshop.catalog.events"

	TypeOrderPlaced        = "OrderPlaced"
	TypeOrderCancelled     = "OrderCancelled"
	TypeOrderStatusUpdated = "OrderStatusUpdated"
	TypePaymentCaptured    = "PaymentCaptured"
	TypePaymentFailed      = "PaymentFailed"
	TypeStockReserved      = "StockReserved"
	TypeStockReleased      = "StockReleased"
)

// BaseEvent is the standard cloud-event header for all Kafka messages.
type BaseEvent struct {
	EventID       string    `json:"event_id"`
	EventType     string    `json:"event_type"`
	AggregateID   string    `json:"aggregate_id"`
	AggregateType string    `json:"aggregate_type"`
	Timestamp     time.Time `json:"timestamp"`
	Version       int       `json:"version"`
}

type OrderPlacedEvent struct {
	BaseEvent
	OrderID         string          `json:"order_id"`
	OrderNumber     string          `json:"order_number"`
	UserID          string          `json:"user_id"`
	GrandTotalCents int64           `json:"grand_total_cents"`
	Currency        string          `json:"currency"`
	Items           []OrderItemData `json:"items"`
}

type OrderItemData struct {
	SKU            string `json:"sku"`
	ProductID      string `json:"product_id"`
	Quantity       int    `json:"quantity"`
	UnitPriceCents int64  `json:"unit_price_cents"`
}

type PaymentCapturedEvent struct {
	BaseEvent
	TransactionID   string `json:"transaction_id"`
	OrderID         string `json:"order_id"`
	AmountCents     int64  `json:"amount_cents"`
	Currency        string `json:"currency"`
	PaymentMethod   string `json:"payment_method"`
	GatewayReference string `json:"gateway_reference"`
}

type StockReservedEvent struct {
	BaseEvent
	ReservationID string          `json:"reservation_id"`
	OrderID       string          `json:"order_id"`
	Items         []OrderItemData `json:"items"`
	ExpiresAt     time.Time       `json:"expires_at"`
}
