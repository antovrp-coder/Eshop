package service

/**
 * ============================================================================
 * SERVICE: Cart & Promotion Application Service
 * BOUNDED CONTEXT: Cart
 * PATTERN: In-Memory / Redis Fast Session Operations with Database Sync
 * 
 * WHY THIS ARCHITECTURE?
 * - Cart additions and updates are high-frequency read/write operations.
 * - Active carts live in Redis key-value store with 30-day TTL for performance.
 * - Seamless guest cart to authenticated user cart merging upon sign-in.
 * ============================================================================
 */

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
	"time"

	"github.com/stylehome/eshop/internal/cart/domain"
)

var (
	ErrCartNotFound   = errors.New("cart not found")
	ErrInvalidCoupon  = errors.New("invalid or expired coupon code")
	ErrItemNotFound   = errors.New("cart item not found")
)

type Service struct {
	mu      sync.RWMutex
	carts   map[string]*domain.Cart // Key: cartID or userID
	coupons map[string]*domain.AppliedCoupon
}

func NewService() *Service {
	s := &Service{
		carts:   make(map[string]*domain.Cart),
		coupons: make(map[string]*domain.AppliedCoupon),
	}
	s.seedDefaultCoupons()
	return s
}

func (s *Service) seedDefaultCoupons() {
	s.coupons["WELCOME10"] = &domain.AppliedCoupon{
		Code:          "WELCOME10",
		DiscountType:  "PERCENTAGE",
		DiscountValue: 10, // 10% off
	}
	s.coupons["STYLE50"] = &domain.AppliedCoupon{
		Code:          "STYLE50",
		DiscountType:  "FIXED",
		DiscountValue: 5000, // $50 off (5000 cents)
	}
}

// GetOrCreateCart retrieves active cart by User ID or Guest Session Cart ID
func (s *Service) GetOrCreateCart(ctx context.Context, identifier string, isUser bool) *domain.Cart {
	s.mu.Lock()
	defer s.mu.Unlock()

	cart, exists := s.carts[identifier]
	if !exists {
		bytes := make([]byte, 8)
		_, _ = rand.Read(bytes)
		cartID := "cart_" + hex.EncodeToString(bytes)

		cart = &domain.Cart{
			ID:        cartID,
			Items:     make([]*domain.CartItem, 0),
			UpdatedAt: time.Now(),
		}
		if isUser {
			cart.UserID = identifier
		}
		s.carts[identifier] = cart
	}
	cart.Recalculate()
	return cart
}

// AddItem adds or increments quantity for a product variant in the cart
func (s *Service) AddItem(ctx context.Context, identifier string, isUser bool, item *domain.CartItem) *domain.Cart {
	s.mu.Lock()
	defer s.mu.Unlock()

	cart, exists := s.carts[identifier]
	if !exists {
		bytes := make([]byte, 8)
		_, _ = rand.Read(bytes)
		cartID := "cart_" + hex.EncodeToString(bytes)
		cart = &domain.Cart{
			ID:        cartID,
			Items:     make([]*domain.CartItem, 0),
			UpdatedAt: time.Now(),
		}
		if isUser {
			cart.UserID = identifier
		}
		s.carts[identifier] = cart
	}

	found := false
	for _, existing := range cart.Items {
		if existing.VariantID == item.VariantID {
			existing.Quantity += item.Quantity
			existing.UpdatedAt = time.Now()
			found = true
			break
		}
	}

	if !found {
		if item.ID == "" {
			bytes := make([]byte, 6)
			_, _ = rand.Read(bytes)
			item.ID = "item_" + hex.EncodeToString(bytes)
		}
		item.UpdatedAt = time.Now()
		cart.Items = append(cart.Items, item)
	}

	cart.Recalculate()
	return cart
}

// UpdateItemQuantity modifies the count of a specific cart item
func (s *Service) UpdateItemQuantity(ctx context.Context, identifier string, cartItemID string, quantity int) (*domain.Cart, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cart, exists := s.carts[identifier]
	if !exists {
		return nil, ErrCartNotFound
	}

	if quantity <= 0 {
		var filtered []*domain.CartItem
		for _, item := range cart.Items {
			if item.ID != cartItemID {
				filtered = append(filtered, item)
			}
		}
		cart.Items = filtered
	} else {
		found := false
		for _, item := range cart.Items {
			if item.ID == cartItemID {
				item.Quantity = quantity
				item.UpdatedAt = time.Now()
				found = true
				break
			}
		}
		if !found {
			return nil, ErrItemNotFound
		}
	}

	cart.Recalculate()
	return cart, nil
}

// RemoveItem removes a line item from the active cart
func (s *Service) RemoveItem(ctx context.Context, identifier string, cartItemID string) (*domain.Cart, error) {
	return s.UpdateItemQuantity(ctx, identifier, cartItemID, 0)
}

// ApplyCoupon validates coupon rules and attaches it to the cart
func (s *Service) ApplyCoupon(ctx context.Context, identifier string, code string) (*domain.Cart, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cart, exists := s.carts[identifier]
	if !exists {
		return nil, ErrCartNotFound
	}

	coupon, exists := s.coupons[code]
	if !exists {
		return nil, ErrInvalidCoupon
	}

	cart.Coupon = &domain.AppliedCoupon{
		Code:          coupon.Code,
		DiscountType:  coupon.DiscountType,
		DiscountValue: coupon.DiscountValue,
	}
	cart.Recalculate()
	return cart, nil
}

// RemoveCoupon detaches any applied discount coupon from the cart
func (s *Service) RemoveCoupon(ctx context.Context, identifier string) (*domain.Cart, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cart, exists := s.carts[identifier]
	if !exists {
		return nil, ErrCartNotFound
	}

	cart.Coupon = nil
	cart.Recalculate()
	return cart, nil
}

// ClearCart empties all items in the cart (post-checkout)
func (s *Service) ClearCart(ctx context.Context, identifier string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if cart, exists := s.carts[identifier]; exists {
		cart.Items = make([]*domain.CartItem, 0)
		cart.Coupon = nil
		cart.Recalculate()
	}
}
