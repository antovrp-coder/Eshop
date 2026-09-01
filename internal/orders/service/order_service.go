package service

/**
 * ============================================================================
 * SERVICE: Orders CQRS Service (Commands & Queries)
 * BOUNDED CONTEXT: Orders
 * PATTERN: Command Query Responsibility Segregation (CQRS)
 * 
 * WHY CQRS SEPARATION?
 * - Commands (PlaceOrder, CancelOrder, UpdateStatus): Heavy transactional consistency,
 *   Saga orchestration, PostgreSQL ACID write, and outbox event publishing.
 * - Queries (GetOrder, ListOrders): Direct read-optimized models served from PostgreSQL
 *   read replicas and cached in Redis.
 * ============================================================================
 */

import (
	"context"
	"errors"
	"sync"

	"github.com/stylehome/eshop/internal/orders/application"
	"github.com/stylehome/eshop/internal/orders/domain"
)

var (
	ErrOrderNotFound = errors.New("order not found")
)

type Service struct {
	mu           sync.RWMutex
	orders       map[string]*domain.Order
	userOrders   map[string][]string // userID -> []orderID
	orchestrator *application.OrderSagaOrchestrator
}

func NewService(orchestrator *application.OrderSagaOrchestrator) *Service {
	return &Service{
		orders:       make(map[string]*domain.Order),
		userOrders:   make(map[string][]string),
		orchestrator: orchestrator,
	}
}

// PlaceOrder executes the CQRS Command via Saga Orchestrator
func (s *Service) PlaceOrder(ctx context.Context, cmd application.PlaceOrderCommand) (*domain.Order, error) {
	order, err := s.orchestrator.Execute(ctx, cmd)
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.orders[order.ID] = order
	s.userOrders[order.UserID] = append(s.userOrders[order.UserID], order.ID)
	return order, nil
}

// CancelOrder executes cancellation command
func (s *Service) CancelOrder(ctx context.Context, orderID string, userID string, reason string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	order, exists := s.orders[orderID]
	if !exists {
		return ErrOrderNotFound
	}

	if order.UserID != userID {
		return errors.New("unauthorized to cancel this order")
	}

	return order.TransitionTo(domain.StatusCancelled)
}

// GetOrderByID provides fast query lookups (CQRS Query Path)
func (s *Service) GetOrderByID(ctx context.Context, orderID string) (*domain.Order, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	order, exists := s.orders[orderID]
	if !exists {
		return nil, ErrOrderNotFound
	}
	return order, nil
}

// ListOrdersByUser retrieves customer order history (CQRS Query Path)
func (s *Service) ListOrdersByUser(ctx context.Context, userID string) []*domain.Order {
	s.mu.RLock()
	defer s.mu.RUnlock()

	orderIDs := s.userOrders[userID]
	var list []*domain.Order
	for i := len(orderIDs) - 1; i >= 0; i-- { // Most recent first
		if ord, exists := s.orders[orderIDs[i]]; exists {
			list = append(list, ord)
		}
	}
	return list
}
