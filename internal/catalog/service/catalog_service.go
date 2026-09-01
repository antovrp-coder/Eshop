package service

/**
 * ============================================================================
 * SERVICE: Catalog Application Service
 * BOUNDED CONTEXT: Catalog
 * TECHNOLOGY: PostgreSQL (`Eshop` db) + Redis Hot Cache
 * ============================================================================
 */

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/stylehome/eshop/internal/catalog/domain"
)

var (
	ErrProductNotFound  = errors.New("product not found")
	ErrCategoryNotFound = errors.New("category not found")
)

type Service struct {
	mu         sync.RWMutex
	products   map[string]*domain.Product   // ID -> Product
	slugMap    map[string]string            // Slug -> ID
	categories map[string]*domain.Category  // Slug -> Category
}

func NewService() *Service {
	s := &Service{
		products:   make(map[string]*domain.Product),
		slugMap:    make(map[string]string),
		categories: make(map[string]*domain.Category),
	}
	s.seedCatalog()
	return s
}

func (s *Service) seedCatalog() {
	// Categories
	cats := []*domain.Category{
		{ID: "cat_living", Name: "Living Room", Slug: "living-room", Description: "Architectural seating, sofas, and coffee tables.", ImageURL: "https://images.unsplash.com/photo-1555041469-a586c61ea9bc?auto=format&fit=crop&w=800&q=80"},
		{ID: "cat_lighting", Name: "Modern Lighting", Slug: "lighting", Description: "Pendant lights, Scandinavian floor lamps, and sconces.", ImageURL: "https://images.unsplash.com/photo-1507473885765-e6ed057f782c?auto=format&fit=crop&w=800&q=80"},
		{ID: "cat_workspace", Name: "Study & Office", Slug: "workspace", Description: "Ergonomic oak desks, chairs, and acoustic storage.", ImageURL: "https://images.unsplash.com/photo-1518455027359-f3f8164ba6bd?auto=format&fit=crop&w=800&q=80"},
		{ID: "cat_decor", Name: "Art & Ceramics", Slug: "art-decor", Description: "Handmade vases, stoneware accents, and tactile textiles.", ImageURL: "https://images.unsplash.com/photo-1612196808214-b8e1d6145a8c?auto=format&fit=crop&w=800&q=80"},
	}
	for _, c := range cats {
		s.categories[c.Slug] = c
	}

	// Curated StyleHome Products
	prods := []*domain.Product{
		{
			ID: "prod_001",
			Title: "Aethelgard Velvet Sectional Sofa",
			Slug: "aethelgard-velvet-sectional-sofa",
			Description: "Deep-seated modular sofa tailored in Italian velvet with solid FSC-certified walnut legs and feather-down wrap cushions.",
			Brand: "StyleHome Atelier",
			CategorySlug: "living-room",
			Tags: []string{"sofa", "velvet", "living", "luxury"},
			BasePriceCents: 249900, // $2,499.00
			Currency: "USD",
			FeaturedImage: "https://images.unsplash.com/photo-1555041469-a586c61ea9bc?auto=format&fit=crop&w=1000&q=80",
			Images: []string{
				"https://images.unsplash.com/photo-1555041469-a586c61ea9bc?auto=format&fit=crop&w=1000&q=80",
				"https://images.unsplash.com/photo-1493663284031-b7e3aefcae8e?auto=format&fit=crop&w=1000&q=80",
			},
			AverageRating: 4.9,
			ReviewCount: 84,
			IsFeatured: true,
			Variants: []domain.ProductVariant{
				{
					ID: "var_001_blu",
					SKU: "VELVET-SOFA-BLU",
					Title: "Midnight Navy / Right Sectional",
					PriceCents: 249900,
					CompareAtPriceCents: 289900,
					InStock: true,
					ImageURL: "https://images.unsplash.com/photo-1555041469-a586c61ea9bc?auto=format&fit=crop&w=1000&q=80",
					Attributes: []domain.ProductAttribute{
						{Name: "Color", Value: "Midnight Navy"},
						{Name: "Configuration", Value: "Right-Facing"},
					},
				},
				{
					ID: "var_001_grn",
					SKU: "VELVET-SOFA-GRN",
					Title: "Forest Emerald / Left Sectional",
					PriceCents: 259900,
					InStock: true,
					ImageURL: "https://images.unsplash.com/photo-1493663284031-b7e3aefcae8e?auto=format&fit=crop&w=1000&q=80",
					Attributes: []domain.ProductAttribute{
						{Name: "Color", Value: "Forest Emerald"},
						{Name: "Configuration", Value: "Left-Facing"},
					},
				},
			},
			CreatedAt: time.Now().Add(-30 * 24 * time.Hour),
		},
		{
			ID: "prod_002",
			Title: "Lumen Minimalist Oak Dining Table",
			Slug: "lumen-minimalist-oak-dining-table",
			Description: "Crafted from single-origin white oak with soft chamfered edges and concealed structural steel braces.",
			Brand: "Nordic Forms",
			CategorySlug: "living-room",
			Tags: []string{"table", "oak", "minimalist", "dining"},
			BasePriceCents: 125000, // $1,250.00
			Currency: "USD",
			FeaturedImage: "https://images.unsplash.com/photo-1577140917170-285929fb55b7?auto=format&fit=crop&w=1000&q=80",
			Images: []string{
				"https://images.unsplash.com/photo-1577140917170-285929fb55b7?auto=format&fit=crop&w=1000&q=80",
			},
			AverageRating: 4.8,
			ReviewCount: 52,
			IsFeatured: true,
			Variants: []domain.ProductVariant{
				{
					ID: "var_002_72",
					SKU: "LUMEN-OAK-01",
					Title: "72 inch / Natural Oak",
					PriceCents: 125000,
					InStock: true,
					ImageURL: "https://images.unsplash.com/photo-1577140917170-285929fb55b7?auto=format&fit=crop&w=1000&q=80",
					Attributes: []domain.ProductAttribute{
						{Name: "Size", Value: "72 inch (Seats 6)"},
						{Name: "Finish", Value: "Natural Matte"},
					},
				},
			},
			CreatedAt: time.Now().Add(-15 * 24 * time.Hour),
		},
		{
			ID: "prod_003",
			Title: "Nordic Brass & Opal Arc Floor Lamp",
			Slug: "nordic-brass-opal-arc-floor-lamp",
			Description: "Hand-spun brushed brass arc with a handblown frosted opal glass globe and solid travertine pedestal base.",
			Brand: "Lumiere Studio",
			CategorySlug: "lighting",
			Tags: []string{"lighting", "lamp", "brass", "scandinavian"},
			BasePriceCents: 42000, // $420.00
			Currency: "USD",
			FeaturedImage: "https://images.unsplash.com/photo-1507473885765-e6ed057f782c?auto=format&fit=crop&w=1000&q=80",
			Images: []string{
				"https://images.unsplash.com/photo-1507473885765-e6ed057f782c?auto=format&fit=crop&w=1000&q=80",
			},
			AverageRating: 4.95,
			ReviewCount: 118,
			IsFeatured: true,
			Variants: []domain.ProductVariant{
				{
					ID: "var_003_std",
					SKU: "NORDIC-FLOOR-LAMP",
					Title: "Brushed Brass / Opal",
					PriceCents: 42000,
					CompareAtPriceCents: 49000,
					InStock: true,
					ImageURL: "https://images.unsplash.com/photo-1507473885765-e6ed057f782c?auto=format&fit=crop&w=1000&q=80",
					Attributes: []domain.ProductAttribute{
						{Name: "Finish", Value: "Brushed Brass"},
					},
				},
			},
			CreatedAt: time.Now().Add(-10 * 24 * time.Hour),
		},
		{
			ID: "prod_004",
			Title: "Wabi-Sabi Stoneware Ceramic Vessel Set",
			Slug: "wabi-sabi-stoneware-ceramic-vessel-set",
			Description: "Trio of sculptural terracotta and grogged stoneware vessels finished in reactive matte glaze.",
			Brand: "Terra Arte",
			CategorySlug: "art-decor",
			Tags: []string{"ceramics", "decor", "vase", "handmade"},
			BasePriceCents: 18500, // $185.00
			Currency: "USD",
			FeaturedImage: "https://images.unsplash.com/photo-1612196808214-b8e1d6145a8c?auto=format&fit=crop&w=1000&q=80",
			Images: []string{
				"https://images.unsplash.com/photo-1612196808214-b8e1d6145a8c?auto=format&fit=crop&w=1000&q=80",
			},
			AverageRating: 4.7,
			ReviewCount: 39,
			IsFeatured: true,
			Variants: []domain.ProductVariant{
				{
					ID: "var_004_set",
					SKU: "CERAMIC-VASE-SET",
					Title: "Set of 3 / Sand Dune",
					PriceCents: 18500,
					InStock: true,
					ImageURL: "https://images.unsplash.com/photo-1612196808214-b8e1d6145a8c?auto=format&fit=crop&w=1000&q=80",
					Attributes: []domain.ProductAttribute{
						{Name: "Color", Value: "Sand Dune"},
					},
				},
			},
			CreatedAt: time.Now().Add(-5 * 24 * time.Hour),
		},
		{
			ID: "prod_005",
			Title: "Kanso Solid Walnut Ergonomic Desk",
			Slug: "kanso-solid-walnut-ergonomic-desk",
			Description: "Dual-motor motorized sit-to-stand desk with a 1.75 inch thick solid American walnut desktop and integrated wire management tray.",
			Brand: "StyleHome Atelier",
			CategorySlug: "workspace",
			Tags: []string{"desk", "workspace", "walnut", "ergonomic"},
			BasePriceCents: 98000, // $980.00
			Currency: "USD",
			FeaturedImage: "https://images.unsplash.com/photo-1518455027359-f3f8164ba6bd?auto=format&fit=crop&w=1000&q=80",
			Images: []string{
				"https://images.unsplash.com/photo-1518455027359-f3f8164ba6bd?auto=format&fit=crop&w=1000&q=80",
			},
			AverageRating: 4.9,
			ReviewCount: 65,
			IsFeatured: true,
			Variants: []domain.ProductVariant{
				{
					ID: "var_005_wht",
					SKU: "MINIMAL-DESK-WHT",
					Title: "60x30 in / Walnut + White Frame",
					PriceCents: 98000,
					InStock: true,
					ImageURL: "https://images.unsplash.com/photo-1518455027359-f3f8164ba6bd?auto=format&fit=crop&w=1000&q=80",
					Attributes: []domain.ProductAttribute{
						{Name: "Size", Value: "60 x 30 inches"},
						{Name: "Frame Color", Value: "Matte White"},
					},
				},
			},
			CreatedAt: time.Now().Add(-20 * 24 * time.Hour),
		},
	}

	for _, p := range prods {
		s.products[p.ID] = p
		s.slugMap[p.Slug] = p.ID
	}
}

func (s *Service) GetProductByID(ctx context.Context, id string) (*domain.Product, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, exists := s.products[id]
	if !exists {
		return nil, ErrProductNotFound
	}
	return p, nil
}

func (s *Service) GetProductBySlug(ctx context.Context, slug string) (*domain.Product, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	id, exists := s.slugMap[slug]
	if !exists {
		return nil, ErrProductNotFound
	}
	return s.products[id], nil
}

func (s *Service) ListCategories(ctx context.Context) []*domain.Category {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var list []*domain.Category
	for _, c := range s.categories {
		list = append(list, c)
	}
	return list
}

func (s *Service) ListAllProducts(ctx context.Context) []*domain.Product {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var list []*domain.Product
	for _, p := range s.products {
		list = append(list, p)
	}
	return list
}
