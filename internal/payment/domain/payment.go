package domain

/**
 * ============================================================================
 * DOMAIN: PAYMENT TRANSACTIONS & LEDGER
 * BOUNDED CONTEXT: Payment
 * PATTERN: DDD Aggregate Root & PCI Tokenized Transactions
 * 
 * CORE ARCHITECTURAL RULES:
 * 1. Monies are strictly represented as integers in minor currency units (Cents).
 * 2. Raw card numbers are NEVER accepted or stored (PCI-DSS Level 1 compliance).
 * 3. Payment lifecycle transitions: PENDING -> AUTHORIZED -> CAPTURED -> REFUNDED / VOIDED.
 * ============================================================================
 */

import (
	"errors"
	"time"
)

var (
	ErrInvalidAmount     = errors.New("amount must be greater than zero")
	ErrTransactionClosed = errors.New("cannot mutate captured/voided transaction")
)

type Status string

const (
	StatusPending    Status = "PENDING"
	StatusAuthorized Status = "AUTHORIZED"
	StatusCaptured   Status = "CAPTURED"
	StatusVoided     Status = "VOIDED"
	StatusRefunded   Status = "REFUNDED"
	StatusFailed     Status = "FAILED"
)

type Money struct {
	AmountInCents int64
	Currency      string
}

type Transaction struct {
	ID                 string
	OrderID            string
	UserID             string
	Amount             Money
	PaymentMethod      string
	GatewayProvider    string
	GatewayReferenceID string
	Status             Status
	IdempotencyKey     string
	FailureReason      string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}
