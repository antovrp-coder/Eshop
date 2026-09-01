package domain

/**
 * ============================================================================
 * DOMAIN: ORDERS AGGREGATE ROOT & STATE MACHINE
 * BOUNDED CONTEXT: Orders
 * PATTERN: DDD Aggregate Root & State Transitions
 * 
 * VALID STATE TRANSITIONS:
 * PENDING_PAYMENT -> PAYMENT_CONFIRMED -> PROCESSING -> SHIPPED -> DELIVERED
 *                 -> CANCELLED
 * PAYMENT_CONFIRMED -> REFUNDED
 * ============================================================================
 */

import (
	"errors"
	"time"
)

var (
	ErrInvalidStateTransition = errors.New("invalid order status transition")
	ErrOrderAlreadyProcessed  = errors.New("cannot cancel or modify order in current status")
)

type OrderStatus string

const (
	StatusPendingPayment   OrderStatus = "PENDING_PAYMENT"
	StatusPaymentConfirmed OrderStatus = "PAYMENT_CONFIRMED"
	StatusProcessing       OrderStatus = "PROCESSING"
	StatusShipped          OrderStatus = "SHIPPED"
	StatusDelivered        OrderStatus = "DELIVERED"
	StatusCancelled        OrderStatus = "CANCELLED"
	StatusRefunded         OrderStatus = "REFUNDED"
)

type Address struct {
	FullName      string `json:"full_name"`
	StreetLine1   string `json:"street_line1"`
	StreetLine2   string `json:"street_line2,omitempty"`
	City          string `json:"city"`
	StateProvince string `json:"state_province"`
	PostalCode    string `json:"postal_code"`
	CountryCode   string `json:"country_code"`
	PhoneNumber   string `json:"phone_number"`
}

type OrderLineItem struct {
	ID             string `json:"id"`
	ProductID      string `json:"product_id"`
	VariantID      string `json:"variant_id"`
	SKU            string `json:"sku"`
	ProductTitle   string `json:"product_title"`
	VariantTitle   string `json:"variant_title,omitempty"`
	UnitPriceCents int64  `json:"unit_price_cents"`
	TotalPriceCents int64 `json:"total_price_cents"`
	Quantity       int    `json:"quantity"`
	ImageURL       string `json:"image_url"`
}

type Order struct {
	ID              string          `json:"id"`
	OrderNumber     string          `json:"order_number"`
	UserID          string          `json:"user_id"`
	Status          OrderStatus     `json:"status"`
	Items           []OrderLineItem `json:"items"`
	ShippingAddress Address         `json:"shipping_address"`
	BillingAddress  Address         `json:"billing_address"`
	SubtotalCents   int64           `json:"subtotal_cents"`
	DiscountCents   int64           `json:"discount_cents"`
	ShippingCents   int64           `json:"shipping_cents"`
	TaxCents        int64           `json:"tax_cents"`
	GrandTotalCents int64           `json:"grand_total_cents"`
	Currency        string          `json:"currency"`
	PaymentMethod   string          `json:"payment_method"`
	ReservationID   string          `json:"reservation_id,omitempty"`
	TransactionID   string          `json:"transaction_id,omitempty"`
	TrackingNumber  string          `json:"tracking_number,omitempty"`
	Carrier         string          `json:"carrier,omitempty"`
	IdempotencyKey  string          `json:"idempotency_key"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

// CanCancel returns true if order is still prior to shipping
func (o *Order) CanCancel() bool {
	return o.Status == StatusPendingPayment || o.Status == StatusPaymentConfirmed || o.Status == StatusProcessing
}

// TransitionTo validates and mutates the order status
func (o *Order) TransitionTo(newStatus OrderStatus) error {
	switch newStatus {
	case StatusPaymentConfirmed:
		if o.Status != StatusPendingPayment {
			return ErrInvalidStateTransition
		}
	case StatusProcessing:
		if o.Status != StatusPaymentConfirmed {
			return ErrInvalidStateTransition
		}
	case StatusShipped:
		if o.Status != StatusProcessing {
			return ErrInvalidStateTransition
		}
	case StatusDelivered:
		if o.Status != StatusShipped {
			return ErrInvalidStateTransition
		}
	case StatusCancelled:
		if !o.CanCancel() {
			return ErrOrderAlreadyProcessed
		}
	case StatusRefunded:
		if o.Status != StatusPaymentConfirmed && o.Status != StatusProcessing && o.Status != StatusDelivered {
			return ErrInvalidStateTransition
		}
	default:
		return ErrInvalidStateTransition
	}

	o.Status = newStatus
	o.UpdatedAt = time.Now()
	return nil
}
