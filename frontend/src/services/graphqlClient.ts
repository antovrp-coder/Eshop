/**
 * ============================================================================
 * SERVICE: StyleHome GraphQL Client & Query Fetcher
 * TECHNOLOGY: Plain Fetch + React Query Integration
 * ============================================================================
 */

export interface ProductAttribute {
  name: string;
  value: string;
}

export interface ProductVariant {
  id: string;
  sku: string;
  title: string;
  price_cents: number;
  compare_at_price_cents?: number;
  attributes: ProductAttribute[];
  in_stock: boolean;
  image_url: string;
}

export interface Product {
  id: string;
  title: string;
  slug: string;
  description: string;
  brand: string;
  category_slug: string;
  tags: string[];
  base_price_cents: number;
  currency: string;
  variants: ProductVariant[];
  images: string[];
  featured_image: string;
  average_rating: number;
  review_count: number;
  is_featured: boolean;
}

export interface FacetBucket {
  key: string;
  count: number;
}

export interface FacetGroup {
  name: string;
  field: string;
  buckets: FacetBucket[];
}

export interface SearchResult {
  products: Product[];
  total_count: number;
  facets: FacetGroup[];
  suggestions: string[];
  total_pages: number;
  current_page: number;
}

export interface CartItem {
  id: string;
  variant_id: string;
  sku: string;
  product_title: string;
  variant_title: string;
  unit_price_cents: number;
  quantity: number;
  image_url: string;
}

export interface AppliedCoupon {
  code: string;
  discount_type: string;
  discount_value: number;
  discount_amount_cents: number;
}

export interface Cart {
  id: string;
  user_id?: string;
  items: CartItem[];
  coupon?: AppliedCoupon;
  subtotal_cents: number;
  discount_cents: number;
  tax_cents: number;
  shipping_cents: number;
  grand_total_cents: number;
}

export interface Order {
  id: string;
  order_number: string;
  status: string;
  items: Array<{
    id: string;
    product_title: string;
    variant_title: string;
    unit_price_cents: number;
    total_price_cents: number;
    quantity: number;
    image_url: string;
  }>;
  subtotal_cents: number;
  discount_cents: number;
  shipping_cents: number;
  tax_cents: number;
  grand_total_cents: number;
  payment_method: string;
  transaction_id: string;
  created_at: string;
}

export interface Category {
  id: string;
  name: string;
  slug: string;
  description: string;
  image_url: string;
}

const GRAPHQL_ENDPOINT = '/graphql';

export async function fetchGraphQL<T = any>(query: string, variables?: Record<string, any>): Promise<T> {
  const token = localStorage.getItem('stylehome_token');
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
  };
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  const response = await fetch(GRAPHQL_ENDPOINT, {
    method: 'POST',
    headers,
    body: JSON.stringify({ query, variables }),
  });

  if (!response.ok) {
    throw new Error(`HTTP network error: ${response.statusText}`);
  }

  const json = await response.json();
  if (json.errors && json.errors.length > 0) {
    throw new Error(json.errors[0].message || 'GraphQL Query Error');
  }

  return json.data;
}

export const api = {
  // Search and Catalog
  async searchProducts(input: {
    query?: string;
    categorySlug?: string;
    brands?: string[];
    minPrice?: number;
    maxPrice?: number;
    sortBy?: string;
    page?: number;
    limit?: number;
  }): Promise<SearchResult> {
    const query = `
      query SearchProducts($input: SearchProductsInput!) {
        searchProducts(input: $input) {
          products
          total_count
          facets
          suggestions
          total_pages
          current_page
        }
      }
    `;
    const data = await fetchGraphQL(query, { input });
    return data.searchProducts;
  },

  async getSuggestions(prefix: string): Promise<string[]> {
    const query = `
      query SearchSuggestions($prefix: String!) {
        searchSuggestions(prefix: $prefix)
      }
    `;
    const data = await fetchGraphQL(query, { prefix });
    return data.searchSuggestions || [];
  },

  async getProductBySlug(slug: string): Promise<Product> {
    const query = `
      query ProductBySlug($slug: String!) {
        productBySlug(slug: $slug)
      }
    `;
    const data = await fetchGraphQL(query, { slug });
    return data.productBySlug;
  },

  async getCategories(): Promise<Category[]> {
    const query = `
      query Categories {
        categories
      }
    `;
    const data = await fetchGraphQL(query);
    return data.categories || [];
  },

  // Cart & Promotions
  async getCart(): Promise<Cart> {
    const query = `
      query GetCart {
        cart
      }
    `;
    const data = await fetchGraphQL(query);
    return data.cart;
  },

  async addToCart(variantId: string, quantity: number = 1): Promise<Cart> {
    const query = `
      mutation AddToCart($input: AddToCartInput!) {
        addToCart(input: $input)
      }
    `;
    const data = await fetchGraphQL(query, { input: { variantId, quantity } });
    return data.addToCart;
  },

  async updateCartItem(cartItemId: string, quantity: number): Promise<Cart> {
    const query = `
      mutation UpdateCartItemQuantity($input: UpdateCartItemInput!) {
        updateCartItemQuantity(input: $input)
      }
    `;
    const data = await fetchGraphQL(query, { input: { cartItemId, quantity } });
    return data.updateCartItemQuantity;
  },

  async applyCoupon(code: string): Promise<Cart> {
    const query = `
      mutation ApplyCoupon($code: String!) {
        applyCoupon(code: $code)
      }
    `;
    const data = await fetchGraphQL(query, { code });
    return data.applyCoupon;
  },

  // Checkout & Orders
  async placeOrder(): Promise<{ success: boolean; order?: Order; errorMessage?: string }> {
    const query = `
      mutation PlaceOrder {
        placeOrder
      }
    `;
    const data = await fetchGraphQL(query);
    return data.placeOrder;
  },

  async getMyOrders(): Promise<Order[]> {
    const query = `
      query MyOrders {
        myOrders
      }
    `;
    const data = await fetchGraphQL(query);
    return data.myOrders?.orders || [];
  },
};

export function formatPrice(cents: number, currency: string = 'USD'): string {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency,
  }).format(cents / 100);
}
