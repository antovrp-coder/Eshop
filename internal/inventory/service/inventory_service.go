package service

/**
 * ============================================================================
 * SERVICE: Inventory gRPC / Application Service
 * BOUNDED CONTEXT: Inventory
 * PATTERN: CQRS Command Handlers & Distributed Reservations
 * 
 * WHY THIS ARCHITECTURE FOR INVENTORY?
 * - Write Operations (ReserveStock, CommitReservation, ReleaseReservation):
 *   Must execute with strict serializability / row-level locks on (sku, warehouse_id)
 *   to avoid race conditions (overselling) during massive flash-sales.
 * - Read Operations (CheckStock):
 *   Can be cached in Redis with short TTL (e.g., 3-5 seconds) for instant response.
 * ============================================================================
 */

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"

	"github.com/stylehome/eshop/internal/inventory/domain"
)

type ReserveItemRequest struct {
	SKU      string
	Quantity int
}

type ReservationResult struct {
	Success       bool
	ReservationID string
	ExpiresAt     time.Time
	OutSKUs       []string
	Error         string
}

type StockInfo struct {
	SKU       string
	Available int
	Reserved  int
	InStock   bool
}

type Service struct {
	mu           sync.RWMutex
	stocks       map[string]*domain.StockLevel
	reservations map[string]*domain.StockReservation
	stopChan     chan struct{}
}

func NewService(ctx context.Context) *Service {
	s := &Service{
		stocks:       make(map[string]*domain.StockLevel),
		reservations: make(map[string]*domain.StockReservation),
		stopChan:     make(chan struct{}),
	}
	s.seedDefaultInventory()
	s.startBackgroundReservationReaper(ctx)
	return s
}

/**
 * ============================================================================
 * CONCURRENCY PATTERN: Background Periodic Reaper (time.Ticker + select)
 * 
 * WHY A BACKGROUND CLEANER GOROUTINE?
 * When customers start checkout, inventory is held for 15 minutes. If they abandon
 * the cart or lose internet connection, this background goroutine automatically
 * identifies expired reservations, releases held stock back to the available pool,
 * and prevents stock from being permanently locked up.
 * ============================================================================
 */
func (s *Service) startBackgroundReservationReaper(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)

	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-s.stopChan:
				ticker.Stop()
				return
			case <-ticker.C:
				s.reapExpiredReservations()
			}
		}
	}()
}

func (s *Service) reapExpiredReservations() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for _, res := range s.reservations {
		if res.Status == domain.ReservationHeld && now.After(res.ExpiresAt) {
			if stock, exists := s.stocks[res.SKU]; exists {
				stock.Reserved -= res.Quantity
				if stock.Reserved < 0 {
					stock.Reserved = 0
				}
				stock.UpdatedAt = now
			}
			res.Status = domain.ReservationExpired
		}
	}
}

// Stop stops the background reaper goroutine gracefully
func (s *Service) Stop() {
	close(s.stopChan)
}

func (s *Service) seedDefaultInventory() {
	// Seed mock inventory for StyleHome products
	initialStock := []struct {
		sku string
		qty int
	}{
		{"LUMEN-OAK-01", 50},
		{"VELVET-SOFA-BLU", 15},
		{"MINIMAL-DESK-WHT", 30},
		{"CERAMIC-VASE-SET", 100},
		{"NORDIC-FLOOR-LAMP", 45},
		{"ARTISAN-RUG-8X10", 20},
	}

	for _, item := range initialStock {
		s.stocks[item.sku] = &domain.StockLevel{
			SKU:            item.sku,
			WarehouseID:    "wh-main-us-east",
			TotalQuantity:  item.qty,
			Reserved:       0,
			AllowBackorder: false,
			UpdatedAt:      time.Now(),
		}
	}
}

/**
 * ============================================================================
 * CONCURRENCY PATTERN: Parallel Multi-SKU Validation (sync.WaitGroup + sync.RWMutex)
 * 
 * WHY PARALLEL STOCK LOOKUP?
 * In multi-item carts (e.g. 10+ items), inspecting physical and reserved levels
 * across multiple warehouse partitions sequentially creates latency additive to cart size.
 * Concurrent worker goroutines process SKU lookups in parallel.
 * ============================================================================
 */
func (s *Service) CheckStock(ctx context.Context, skus []string) map[string]StockInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]StockInfo)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, sku := range skus {
		wg.Add(1)
		go func(skuKey string) {
			defer wg.Done()
			var info StockInfo
			if stock, exists := s.stocks[skuKey]; exists {
				avail := stock.Available()
				info = StockInfo{
					SKU:       skuKey,
					Available: avail,
					Reserved:  stock.Reserved,
					InStock:   avail > 0,
				}
			} else {
				info = StockInfo{SKU: skuKey, Available: 0, Reserved: 0, InStock: false}
			}

			mu.Lock()
			result[skuKey] = info
			mu.Unlock()
		}(sku)
	}

	wg.Wait()
	return result
}

// ReserveStock executes transactional reservation with TTL (CQRS Command Path)
func (s *Service) ReserveStock(ctx context.Context, orderID string, items []ReserveItemRequest, holdSeconds int) (*ReservationResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if holdSeconds <= 0 {
		holdSeconds = 900 // Default: 15 minutes
	}

	var outSKUs []string
	for _, item := range items {
		stock, exists := s.stocks[item.SKU]
		if !exists || stock.Available() < item.Quantity {
			outSKUs = append(outSKUs, item.SKU)
		}
	}

	if len(outSKUs) > 0 {
		return &ReservationResult{
			Success: false,
			OutSKUs: outSKUs,
			Error:   "One or more items are out of stock",
		}, nil
	}

	// Generate Reservation Token
	bytes := make([]byte, 8)
	_, _ = rand.Read(bytes)
	reservationID := "res_" + hex.EncodeToString(bytes)
	expiresAt := time.Now().Add(time.Duration(holdSeconds) * time.Second)

	// Apply reservation row hold
	for _, item := range items {
		stock := s.stocks[item.SKU]
		stock.Reserved += item.Quantity
		stock.UpdatedAt = time.Now()

		s.reservations[reservationID] = &domain.StockReservation{
			ID:          reservationID,
			OrderID:     orderID,
			SKU:         item.SKU,
			WarehouseID: stock.WarehouseID,
			Quantity:    item.Quantity,
			Status:      domain.ReservationHeld,
			ExpiresAt:   expiresAt,
			CreatedAt:   time.Now(),
		}
	}

	return &ReservationResult{
		Success:       true,
		ReservationID: reservationID,
		ExpiresAt:     expiresAt,
	}, nil
}

// CommitReservation transitions reservation from HELD to COMMITTED (Saga Step: Payment Success)
func (s *Service) CommitReservation(ctx context.Context, reservationID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	res, exists := s.reservations[reservationID]
	if !exists {
		return domain.ErrReservationNotFound
	}

	if stock, exists := s.stocks[res.SKU]; exists {
		stock.TotalQuantity -= res.Quantity
		stock.Reserved -= res.Quantity
		stock.UpdatedAt = time.Now()
	}

	res.Status = domain.ReservationCommitted
	return nil
}

// ReleaseReservation executes compensating transaction if payment fails or order expires
func (s *Service) ReleaseReservation(ctx context.Context, reservationID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	res, exists := s.reservations[reservationID]
	if !exists || res.Status != domain.ReservationHeld {
		return nil // Already released or non-existent
	}

	if stock, exists := s.stocks[res.SKU]; exists {
		stock.Reserved -= res.Quantity
		if stock.Reserved < 0 {
			stock.Reserved = 0
		}
		stock.UpdatedAt = time.Now()
	}

	res.Status = domain.ReservationReleased
	return nil
}
