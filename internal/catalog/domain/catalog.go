package domain

/**
 * ============================================================================
 * DOMAIN: CATALOG & MERCHANDISING
 * BOUNDED CONTEXT: Catalog
 * PATTERN: DDD Aggregate Root
 * ============================================================================
 */

import "time"

type ProductAttribute struct {
	Name  string `json:"name"`  // e.g., "Color", "Material", "Dimensions"
	Value string `json:"value"` // e.g., "Walnut", "Solid Oak", "W: 120cm D: 60cm"
}

type ProductVariant struct {
	ID                  string             `json:"id"`
	SKU                 string             `json:"sku"`
	Title               string             `json:"title"`
	PriceCents          int64              `json:"price_cents"`
	CompareAtPriceCents int64              `json:"compare_at_price_cents,omitempty"`
	Attributes          []ProductAttribute `json:"attributes"`
	InStock             bool               `json:"in_stock"`
	ImageURL            string             `json:"image_url"`
}

type Category struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	ParentID    string `json:"parent_id,omitempty"`
	ImageURL    string `json:"image_url"`
}

type Product struct {
	ID            string            `json:"id"`
	Title         string            `json:"title"`
	Slug          string            `json:"slug"`
	Description   string            `json:"description"`
	Brand         string            `json:"brand"`
	CategorySlug  string            `json:"category_slug"`
	Tags          []string          `json:"tags"`
	BasePriceCents int64            `json:"base_price_cents"`
	Currency      string            `json:"currency"`
	Variants      []ProductVariant  `json:"variants"`
	Images        []string          `json:"images"`
	FeaturedImage string            `json:"featured_image"`
	AverageRating float64           `json:"average_rating"`
	ReviewCount   int               `json:"review_count"`
	IsFeatured    bool              `json:"is_featured"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
}
