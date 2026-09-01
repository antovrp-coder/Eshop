-- ============================================================================
-- STYLEHOME ESHOP DATABASE MIGRATION - POSTGRESQL
-- DATABASE: Eshop
-- DOMAINS: Auth, Catalog, Inventory, Cart, Orders, Payments, Outbox Events
-- ============================================================================

-- Create isolated schemas for Domain-Driven Design bounded contexts
CREATE SCHEMA IF NOT EXISTS auth_schema;
CREATE SCHEMA IF NOT EXISTS catalog_schema;
CREATE SCHEMA IF NOT EXISTS inventory_schema;
CREATE SCHEMA IF NOT EXISTS cart_schema;
CREATE SCHEMA IF NOT EXISTS order_schema;
CREATE SCHEMA IF NOT EXISTS payment_schema;
CREATE SCHEMA IF NOT EXISTS promotion_schema;

-- ============================================================================
-- 1. AUTH & USER DOMAIN
-- ============================================================================

CREATE TABLE IF NOT EXISTS auth_schema.users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL, -- Argon2id / bcrypt hashed
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    phone_number VARCHAR(30),
    role VARCHAR(50) NOT NULL DEFAULT 'CUSTOMER', -- 'CUSTOMER', 'ADMIN', 'VENDOR', 'SUPPORT'
    is_email_verified BOOLEAN DEFAULT FALSE,
    is_phone_verified BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS auth_schema.addresses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES auth_schema.users(id) ON DELETE CASCADE,
    full_name VARCHAR(150) NOT NULL,
    street_line1 VARCHAR(255) NOT NULL,
    street_line2 VARCHAR(255),
    city VARCHAR(100) NOT NULL,
    state_province VARCHAR(100) NOT NULL,
    postal_code VARCHAR(30) NOT NULL,
    country_code VARCHAR(2) NOT NULL, -- ISO Alpha-2
    phone_number VARCHAR(30) NOT NULL,
    is_default_shipping BOOLEAN DEFAULT FALSE,
    is_default_billing BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS auth_schema.refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES auth_schema.users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================================
-- 2. CATALOG DOMAIN
-- ============================================================================

CREATE TABLE IF NOT EXISTS catalog_schema.categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(150) NOT NULL,
    slug VARCHAR(150) UNIQUE NOT NULL,
    description TEXT,
    parent_id UUID REFERENCES catalog_schema.categories(id) ON DELETE SET NULL,
    image_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS catalog_schema.products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    category_id UUID NOT NULL REFERENCES catalog_schema.categories(id),
    title VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    description TEXT NOT NULL,
    brand VARCHAR(100) NOT NULL,
    base_price_cents BIGINT NOT NULL, -- Stored in minor currency units (e.g. cents) to prevent float rounding errors
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    is_featured BOOLEAN DEFAULT FALSE,
    average_rating NUMERIC(3, 2) DEFAULT 0.00,
    review_count INT DEFAULT 0,
    seo_title VARCHAR(255),
    seo_description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS catalog_schema.product_variants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES catalog_schema.products(id) ON DELETE CASCADE,
    sku VARCHAR(100) UNIQUE NOT NULL,
    title VARCHAR(150) NOT NULL,
    price_cents BIGINT NOT NULL,
    compare_at_price_cents BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS catalog_schema.variant_attributes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    variant_id UUID NOT NULL REFERENCES catalog_schema.product_variants(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,  -- e.g. 'Color', 'Size', 'Material'
    value VARCHAR(150) NOT NULL  -- e.g. 'Navy Blue', 'King', 'Solid Teak'
);

CREATE TABLE IF NOT EXISTS catalog_schema.product_images (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES catalog_schema.products(id) ON DELETE CASCADE,
    variant_id UUID REFERENCES catalog_schema.product_variants(id) ON DELETE SET NULL,
    url TEXT NOT NULL,
    alt_text VARCHAR(255),
    is_featured BOOLEAN DEFAULT FALSE,
    sort_order INT DEFAULT 0
);

-- ============================================================================
-- 3. INVENTORY & STOCK DOMAIN (CQRS Optimized)
-- ============================================================================

CREATE TABLE IF NOT EXISTS inventory_schema.warehouses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(150) NOT NULL,
    country_code VARCHAR(2) NOT NULL,
    city VARCHAR(100) NOT NULL,
    is_active BOOLEAN DEFAULT TRUE
);

CREATE TABLE IF NOT EXISTS inventory_schema.stock_levels (
    sku VARCHAR(100) NOT NULL,
    warehouse_id UUID NOT NULL REFERENCES inventory_schema.warehouses(id),
    quantity INT NOT NULL DEFAULT 0,
    reserved INT NOT NULL DEFAULT 0,
    allow_backorder BOOLEAN DEFAULT FALSE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (sku, warehouse_id),
    CONSTRAINT chk_positive_quantity CHECK (quantity >= 0),
    CONSTRAINT chk_valid_reservation CHECK (reserved >= 0 AND reserved <= quantity)
);

CREATE TABLE IF NOT EXISTS inventory_schema.stock_reservations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL,
    sku VARCHAR(100) NOT NULL,
    warehouse_id UUID NOT NULL REFERENCES inventory_schema.warehouses(id),
    quantity INT NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'HELD', -- 'HELD', 'COMMITTED', 'RELEASED', 'EXPIRED'
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================================
-- 4. CART & PROMOTION DOMAIN
-- ============================================================================

CREATE TABLE IF NOT EXISTS cart_schema.carts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID UNIQUE, -- Null for guest session carts
    coupon_code VARCHAR(50),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS cart_schema.cart_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cart_id UUID NOT NULL REFERENCES cart_schema.carts(id) ON DELETE CASCADE,
    variant_id UUID NOT NULL,
    sku VARCHAR(100) NOT NULL,
    quantity INT NOT NULL CHECK (quantity > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (cart_id, variant_id)
);

CREATE TABLE IF NOT EXISTS promotion_schema.coupons (
    code VARCHAR(50) PRIMARY KEY,
    description TEXT,
    discount_type VARCHAR(20) NOT NULL, -- 'PERCENTAGE', 'FIXED_AMOUNT'
    discount_value INT NOT NULL, -- Percentage (e.g. 15 for 15%) or Cents (e.g. 2000 for $20)
    min_order_cents BIGINT DEFAULT 0,
    max_discount_cents BIGINT,
    starts_at TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    usage_limit INT,
    usage_count INT DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE
);

-- ============================================================================
-- 5. ORDERS DOMAIN (CQRS Write Model + Outbox Pattern)
-- ============================================================================

CREATE TABLE IF NOT EXISTS order_schema.orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_number VARCHAR(50) UNIQUE NOT NULL,
    user_id UUID NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'PENDING_PAYMENT',
    -- 'PENDING_PAYMENT', 'PAYMENT_CONFIRMED', 'PROCESSING', 'SHIPPED', 'DELIVERED', 'CANCELLED', 'REFUNDED'
    subtotal_cents BIGINT NOT NULL,
    discount_cents BIGINT NOT NULL DEFAULT 0,
    shipping_cents BIGINT NOT NULL DEFAULT 0,
    tax_cents BIGINT NOT NULL DEFAULT 0,
    grand_total_cents BIGINT NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    shipping_address JSONB NOT NULL, -- Snapshot of address at time of purchase
    billing_address JSONB NOT NULL,
    payment_method VARCHAR(50) NOT NULL,
    reservation_id UUID,
    tracking_number VARCHAR(100),
    carrier VARCHAR(100),
    idempotency_key VARCHAR(255) UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS order_schema.order_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES order_schema.orders(id) ON DELETE CASCADE,
    product_id UUID NOT NULL,
    variant_id UUID NOT NULL,
    sku VARCHAR(100) NOT NULL,
    title VARCHAR(255) NOT NULL,
    variant_title VARCHAR(150),
    unit_price_cents BIGINT NOT NULL,
    total_price_cents BIGINT NOT NULL,
    quantity INT NOT NULL,
    image_url TEXT
);

-- ============================================================================
-- 6. PAYMENT DOMAIN (PCI Tokenization & Ledger)
-- ============================================================================

CREATE TABLE IF NOT EXISTS payment_schema.transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL,
    user_id UUID NOT NULL,
    amount_cents BIGINT NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    payment_method VARCHAR(50) NOT NULL,
    gateway_provider VARCHAR(50) NOT NULL, -- 'STRIPE', 'ADYEN', 'PAYPAL'
    gateway_transaction_id VARCHAR(255),
    status VARCHAR(50) NOT NULL, -- 'PENDING', 'AUTHORIZED', 'CAPTURED', 'VOIDED', 'REFUNDED', 'FAILED'
    idempotency_key VARCHAR(255) UNIQUE NOT NULL,
    failure_code VARCHAR(100),
    failure_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================================
-- 7. TRANSACTIONAL OUTBOX TABLES FOR KAFKA RELIABILITY
-- The Outbox Pattern guarantees that domain events are committed atomically
-- with the local PostgreSQL database transaction, ensuring Zero Lost Events.
-- ============================================================================

CREATE TABLE IF NOT EXISTS order_schema.outbox_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    aggregate_type VARCHAR(100) NOT NULL, -- 'ORDER'
    aggregate_id VARCHAR(255) NOT NULL,   -- order_id
    event_type VARCHAR(100) NOT NULL,     -- 'OrderPlaced', 'OrderCancelled', 'OrderStatusUpdated'
    payload JSONB NOT NULL,
    published BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS payment_schema.outbox_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    aggregate_type VARCHAR(100) NOT NULL, -- 'PAYMENT'
    aggregate_id VARCHAR(255) NOT NULL,   -- transaction_id
    event_type VARCHAR(100) NOT NULL,     -- 'PaymentAuthorized', 'PaymentCaptured', 'PaymentFailed'
    payload JSONB NOT NULL,
    published BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indices for high performance queries
CREATE INDEX IF NOT EXISTS idx_orders_user_id ON order_schema.orders(user_id);
CREATE INDEX IF NOT EXISTS idx_orders_status ON order_schema.orders(status);
CREATE INDEX IF NOT EXISTS idx_order_outbox_unpublished ON order_schema.outbox_events(published) WHERE published = FALSE;
CREATE INDEX IF NOT EXISTS idx_payment_outbox_unpublished ON payment_schema.outbox_events(published) WHERE published = FALSE;
CREATE INDEX IF NOT EXISTS idx_catalog_products_category ON catalog_schema.products(category_id);
CREATE INDEX IF NOT EXISTS idx_catalog_products_brand ON catalog_schema.products(brand);
CREATE INDEX IF NOT EXISTS idx_inventory_sku ON inventory_schema.stock_levels(sku);
