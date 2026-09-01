package domain

/**
 * ============================================================================
 * DOMAIN: INVENTORY & STOCK LEDGER
 * BOUNDED CONTEXT: Inventory
 * PATTERN: DDD Aggregate Root & Invariant Protection
 * 
 * CORE INVARIANTS:
 * 1. An SKU's reserved quantity must never exceed available physical stock
 *    unless backorders are explicitly enabled.
 * 2. Stock reservations automatically expire after a configurable TTL (e.g. 15 mins)
 *    to prevent inventory lockup during abandoned checkout carts.
 * ============================================================================
 */

import (
	"errors"
	"time"
)

var (
	ErrInsufficientStock   = errors.New("insufficient stock for requested SKU")
	ErrReservationNotFound = errors.New("reservation ID not found")
	ErrReservationExpired  = errors.New("reservation has expired")
)

type ReservationStatus string

const (
	ReservationHeld      ReservationStatus = "HELD"
	ReservationCommitted ReservationStatus = "COMMITTED"
	ReservationReleased  ReservationStatus = "RELEASED"
	ReservationExpired   ReservationStatus = "EXPIRED"
)

type StockLevel struct {
	SKU            string
	WarehouseID    string
	TotalQuantity  int
	Reserved       int
	AllowBackorder bool
	UpdatedAt      time.Time
}

func (s *StockLevel) Available() int {
	return s.TotalQuantity - s.Reserved
}

type StockReservation struct {
	ID          string
	OrderID     string
	SKU         string
	WarehouseID string
	Quantity    int
	Status      ReservationStatus
	ExpiresAt   time.Time
	CreatedAt   time.Time
}

func (r *StockReservation) IsActive() bool {
	return r.Status == ReservationHeld && time.Now().Before(r.ExpiresAt)
}
