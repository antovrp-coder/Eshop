package domain

/**
 * ============================================================================
 * DOMAIN: SHOPPING CART & PROMOTIONS
 * BOUNDED CONTEXT: Cart
 * PATTERN: DDD Aggregate Root
 * ============================================================================
 */

import "time"

type CartItem struct {
	ID             string    `json:"id"`
	VariantID      string    `json:"variant_id"`
	SKU            string    `json:"sku"`
	ProductTitle   string    `json:"product_title"`
	VariantTitle   string    `json:"variant_title"`
	UnitPriceCents int64     `json:"unit_price_cents"`
	Quantity       int       `json:"quantity"`
	ImageURL       string    `json:"image_url"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (ci *CartItem) TotalPriceCents() int64 {
	return ci.UnitPriceCents * int64(ci.Quantity)
}

type AppliedCoupon struct {
	Code           string `json:"code"`
	DiscountType   string `json:"discount_type"` // 'PERCENTAGE', 'FIXED'
	DiscountValue  int    `json:"discount_value"`
	DiscountAmount int64  `json:"discount_amount_cents"`
}

type Cart struct {
	ID             string        `json:"id"`
	UserID         string        `json:"user_id,omitempty"` // Empty for anonymous guest session carts
	Items          []*CartItem   `json:"items"`
	Coupon         *AppliedCoupon `json:"coupon,omitempty"`
	SubtotalCents  int64         `json:"subtotal_cents"`
	DiscountCents  int64         `json:"discount_cents"`
	TaxCents       int64         `json:"tax_cents"`
	ShippingCents  int64         `json:"shipping_cents"`
	GrandTotalCents int64        `json:"grand_total_cents"`
	UpdatedAt      time.Time     `json:"updated_at"`
}

func (c *Cart) Recalculate() {
	var subtotal int64
	for _, item := range c.Items {
		subtotal += item.TotalPriceCents()
	}
	c.SubtotalCents = subtotal

	var discount int64
	if c.Coupon != nil {
		if c.Coupon.DiscountType == "PERCENTAGE" {
			discount = (subtotal * int64(c.Coupon.DiscountValue)) / 100
		} else {
			discount = int64(c.Coupon.DiscountValue)
		}
		if discount > subtotal {
			discount = subtotal
		}
		c.Coupon.DiscountAmount = discount
	}
	c.DiscountCents = discount

	// Shipping: Free if subtotal after discount > $100 (10000 cents), else $10 (1000 cents)
	taxable := subtotal - discount
	if taxable > 10000 || taxable == 0 {
		c.ShippingCents = 0
	} else {
		c.ShippingCents = 1000
	}

	// Standard 8% estimated tax
	c.TaxCents = (taxable * 8) / 100
	c.GrandTotalCents = taxable + c.ShippingCents + c.TaxCents
	c.UpdatedAt = time.Now()
}
