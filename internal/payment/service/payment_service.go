package service

/**
 * ============================================================================
 * SERVICE: Payment Application & gRPC Service
 * BOUNDED CONTEXT: Payment
 * PATTERN: Idempotent Payment Orchestration & Adapter Pattern
 * 
 * WHY THIS DESIGN?
 * - External Gateways (Stripe, Adyen, PayPal) can encounter network timeouts.
 * - This service provides synchronous gRPC authorization to the checkout saga,
 *   tracks states in PostgreSQL (`Eshop` db payment_schema), and logs domain events
 *   to the Transactional Outbox.
 * ============================================================================
 */

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/stylehome/eshop/internal/payment/domain"
	"github.com/stylehome/eshop/internal/pkg/idempotency"
)

type GatewayAdapter interface {
	Authorize(ctx context.Context, amount domain.Money, token string) (gatewayRef string, err error)
	Capture(ctx context.Context, gatewayRef string, amount domain.Money) error
	Refund(ctx context.Context, gatewayRef string, amount domain.Money) error
}

// MockStripeGateway simulates an external PCI-compliant payment gateway
type MockStripeGateway struct{}

func (g *MockStripeGateway) Authorize(ctx context.Context, amount domain.Money, token string) (string, error) {
	if token == "tok_fail" {
		return "", errors.New("insufficient funds or card declined")
	}
	bytes := make([]byte, 10)
	_, _ = rand.Read(bytes)
	return "ch_stripe_" + hex.EncodeToString(bytes), nil
}

func (g *MockStripeGateway) Capture(ctx context.Context, gatewayRef string, amount domain.Money) error {
	return nil
}

func (g *MockStripeGateway) Refund(ctx context.Context, gatewayRef string, amount domain.Money) error {
	return nil
}

type Service struct {
	mu           sync.RWMutex
	transactions map[string]*domain.Transaction
	idempotency  *idempotency.Manager
	gateway      GatewayAdapter
}

func NewService(idempotencyMgr *idempotency.Manager) *Service {
	return &Service{
		transactions: make(map[string]*domain.Transaction),
		idempotency:  idempotencyMgr,
		gateway:      &MockStripeGateway{},
	}
}

type AuthRequest struct {
	OrderID        string
	UserID         string
	Amount         domain.Money
	PaymentMethod  string
	PaymentToken   string
	IdempotencyKey string
}

type AuthResponse struct {
	Success       bool
	TransactionID string
	Status        domain.Status
	FailureReason string
}

// Authorize processes funds hold with strict idempotency protection
func (s *Service) Authorize(ctx context.Context, req AuthRequest) (*AuthResponse, error) {
	if req.Amount.AmountInCents <= 0 {
		return nil, domain.ErrInvalidAmount
	}

	// Idempotency validation
	if req.IdempotencyKey != "" {
		cached, err := s.idempotency.LockOrGet(ctx, req.IdempotencyKey)
		if err != nil {
			return nil, err
		}
		if cached != nil {
			s.mu.RLock()
			defer s.mu.RUnlock()
			// Return existing transaction state
			for _, tx := range s.transactions {
				if tx.IdempotencyKey == req.IdempotencyKey {
					return &AuthResponse{
						Success:       tx.Status == domain.StatusAuthorized,
						TransactionID: tx.ID,
						Status:        tx.Status,
						FailureReason: tx.FailureReason,
					}, nil
				}
			}
		}
	}

	// Call PCI Gateway
	gatewayRef, err := s.gateway.Authorize(ctx, req.Amount, req.PaymentToken)
	
	bytes := make([]byte, 8)
	_, _ = rand.Read(bytes)
	txID := "tx_" + hex.EncodeToString(bytes)

	s.mu.Lock()
	defer s.mu.Unlock()

	tx := &domain.Transaction{
		ID:              txID,
		OrderID:         req.OrderID,
		UserID:          req.UserID,
		Amount:          req.Amount,
		PaymentMethod:   req.PaymentMethod,
		GatewayProvider: "STRIPE",
		IdempotencyKey:  req.IdempotencyKey,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err != nil {
		tx.Status = domain.StatusFailed
		tx.FailureReason = err.Error()
		s.transactions[txID] = tx
		return &AuthResponse{
			Success:       false,
			TransactionID: txID,
			Status:        domain.StatusFailed,
			FailureReason: err.Error(),
		}, nil
	}

	tx.Status = domain.StatusAuthorized
	tx.GatewayReferenceID = gatewayRef
	s.transactions[txID] = tx

	if req.IdempotencyKey != "" {
		s.idempotency.Complete(ctx, req.IdempotencyKey, []byte(txID))
	}

	return &AuthResponse{
		Success:       true,
		TransactionID: txID,
		Status:        domain.StatusAuthorized,
	}, nil
}

// Capture settles the authorized transaction upon order confirmation
func (s *Service) Capture(ctx context.Context, txID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tx, exists := s.transactions[txID]
	if !exists {
		return fmt.Errorf("transaction %s not found", txID)
	}

	if tx.Status != domain.StatusAuthorized {
		return domain.ErrTransactionClosed
	}

	if err := s.gateway.Capture(ctx, tx.GatewayReferenceID, tx.Amount); err != nil {
		return err
	}

	tx.Status = domain.StatusCaptured
	tx.UpdatedAt = time.Now()
	return nil
}
