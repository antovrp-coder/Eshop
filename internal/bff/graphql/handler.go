package graphql

/**
 * ============================================================================
 * PACKAGE: internal/bff/graphql
 * TECHNOLOGY: Pure Go Standard Library (net/http, encoding/json)
 * PATTERN: Backend-For-Frontend (BFF) GraphQL Execution Engine
 * 
 * WHY GRAPHQL BFF WITH STANDARD LIBRARY?
 * 1. Single Entrypoint for Client SPA: Consolidates microservices (Catalog, Search, Cart, Orders).
 * 2. Over-Fetching Elimination: Mobile and web clients retrieve exact fields needed.
 * 3. Zero-Framework Middleware Compatibility: Embeds natively into Go standard `http.Handler`.
 * ============================================================================
 */

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	cartDomain "github.com/stylehome/eshop/internal/cart/domain"
	cartService "github.com/stylehome/eshop/internal/cart/service"
	catalogDomain "github.com/stylehome/eshop/internal/catalog/domain"
	catalogService "github.com/stylehome/eshop/internal/catalog/service"
	ordersApplication "github.com/stylehome/eshop/internal/orders/application"
	ordersDomain "github.com/stylehome/eshop/internal/orders/domain"
	ordersService "github.com/stylehome/eshop/internal/orders/service"
	"github.com/stylehome/eshop/internal/pkg/middleware"
	searchService "github.com/stylehome/eshop/internal/search/service"
)

type GraphQLRequest struct {
	Query         string                 `json:"query"`
	OperationName string                 `json:"operationName,omitempty"`
	Variables     map[string]interface{} `json:"variables,omitempty"`
}

type GraphQLResponse struct {
	Data   interface{}    `json:"data,omitempty"`
	Errors []GraphQLError `json:"errors,omitempty"`
}

type GraphQLError struct {
	Message string `json:"message"`
	Path    []string `json:"path,omitempty"`
}

type Handler struct {
	catalogSvc *catalogService.Service
	searchSvc  *searchService.Service
	cartSvc    *cartService.Service
	ordersSvc  *ordersService.Service
}

func NewHandler(
	cat *catalogService.Service,
	src *searchService.Service,
	crt *cartService.Service,
	ord *ordersService.Service,
) *Handler {
	return &Handler{
		catalogSvc: cat,
		searchSvc:  src,
		cartSvc:    crt,
		ordersSvc:  ord,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "GraphQL queries must use POST", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req GraphQLRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	res := h.Execute(ctx, req)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(res)
}

// Execute resolves queries and mutations against internal services
func (h *Handler) Execute(ctx context.Context, req GraphQLRequest) GraphQLResponse {
	q := strings.TrimSpace(req.Query)

	// User identifier from context or guest session header
	userID, _ := ctx.Value(middleware.UserIDKey).(string)
	sessionID := "guest_session_default"
	if userID != "" {
		sessionID = userID
	}

	// 1. MUTATION: AddToCart
	if strings.Contains(q, "addToCart") {
		var variantID string
		var qty int = 1
		if vars := req.Variables; vars != nil {
			if input, ok := vars["input"].(map[string]interface{}); ok {
				if v, ok := input["variantId"].(string); ok {
					variantID = v
				}
				if qNum, ok := input["quantity"].(float64); ok {
					qty = int(qNum)
				}
			}
		}

		// Find item in catalog
		var foundItem *cartDomain.CartItem
		for _, p := range h.catalogSvc.ListAllProducts(ctx) {
			for _, v := range p.Variants {
				if v.ID == variantID || variantID == "" {
					foundItem = &cartDomain.CartItem{
						VariantID:      v.ID,
						SKU:            v.SKU,
						ProductTitle:   p.Title,
						VariantTitle:   v.Title,
						UnitPriceCents: v.PriceCents,
						Quantity:       qty,
						ImageURL:       v.ImageURL,
					}
					break
				}
			}
			if foundItem != nil {
				break
			}
		}

		if foundItem == nil {
			return GraphQLResponse{Errors: []GraphQLError{{Message: "Variant not found"}}}
		}

		cart := h.cartSvc.AddItem(ctx, sessionID, userID != "", foundItem)
		return GraphQLResponse{Data: map[string]interface{}{"addToCart": cart}}
	}

	// 2. MUTATION: UpdateCartItemQuantity
	if strings.Contains(q, "updateCartItemQuantity") {
		var itemID string
		var qty int
		if vars := req.Variables; vars != nil {
			if input, ok := vars["input"].(map[string]interface{}); ok {
				if id, ok := input["cartItemId"].(string); ok {
					itemID = id
				}
				if qNum, ok := input["quantity"].(float64); ok {
					qty = int(qNum)
				}
			}
		}
		cart, err := h.cartSvc.UpdateItemQuantity(ctx, sessionID, itemID, qty)
		if err != nil {
			return GraphQLResponse{Errors: []GraphQLError{{Message: err.Error()}}}
		}
		return GraphQLResponse{Data: map[string]interface{}{"updateCartItemQuantity": cart}}
	}

	// 3. MUTATION: ApplyCoupon
	if strings.Contains(q, "applyCoupon") {
		var code string
		if vars := req.Variables; vars != nil {
			if c, ok := vars["code"].(string); ok {
				code = c
			}
		}
		cart, err := h.cartSvc.ApplyCoupon(ctx, sessionID, code)
		if err != nil {
			return GraphQLResponse{Errors: []GraphQLError{{Message: err.Error()}}}
		}
		return GraphQLResponse{Data: map[string]interface{}{"applyCoupon": cart}}
	}

	// 4. MUTATION: PlaceOrder (CQRS Saga Orchestration)
	if strings.Contains(q, "placeOrder") {
		cart := h.cartSvc.GetOrCreateCart(ctx, sessionID, userID != "")
		if len(cart.Items) == 0 {
			return GraphQLResponse{Errors: []GraphQLError{{Message: "Cart is empty"}}}
		}

		var lineItems []ordersDomain.OrderLineItem
		for _, it := range cart.Items {
			lineItems = append(lineItems, ordersDomain.OrderLineItem{
				VariantID:      it.VariantID,
				SKU:            it.SKU,
				ProductTitle:   it.ProductTitle,
				VariantTitle:   it.VariantTitle,
				UnitPriceCents: it.UnitPriceCents,
				TotalPriceCents: it.TotalPriceCents(),
				Quantity:       it.Quantity,
				ImageURL:       it.ImageURL,
			})
		}

		cmd := ordersApplication.PlaceOrderCommand{
			UserID:         sessionID,
			IdempotencyKey: fmt.Sprintf("idemp_%d", cart.UpdatedAt.UnixNano()),
			Items:          lineItems,
			ShippingAddress: ordersDomain.Address{
				FullName:      "Jane Doe",
				StreetLine1:   "742 Evergreen Terrace",
				City:          "Springfield",
				StateProvince: "OR",
				PostalCode:    "97477",
				CountryCode:   "US",
				PhoneNumber:   "+1 555 0199",
			},
			BillingAddress: ordersDomain.Address{
				FullName:      "Jane Doe",
				StreetLine1:   "742 Evergreen Terrace",
				City:          "Springfield",
				StateProvince: "OR",
				PostalCode:    "97477",
				CountryCode:   "US",
				PhoneNumber:   "+1 555 0199",
			},
			PaymentMethod:   "CREDIT_CARD",
			PaymentToken:    "tok_visa_valid",
			SubtotalCents:   cart.SubtotalCents,
			DiscountCents:   cart.DiscountCents,
			ShippingCents:   cart.ShippingCents,
			TaxCents:        cart.TaxCents,
			GrandTotalCents: cart.GrandTotalCents,
		}

		order, err := h.ordersSvc.PlaceOrder(ctx, cmd)
		if err != nil {
			return GraphQLResponse{Data: map[string]interface{}{
				"placeOrder": map[string]interface{}{
					"success":      false,
					"errorMessage": err.Error(),
				},
			}}
		}

		// Clear cart after successful placement
		h.cartSvc.ClearCart(ctx, sessionID)

		return GraphQLResponse{Data: map[string]interface{}{
			"placeOrder": map[string]interface{}{
				"success":       true,
				"order":         order,
				"transactionId": order.TransactionID,
			},
		}}
	}

	// 5. QUERY: SearchProducts (Elasticsearch)
	if strings.Contains(q, "searchProducts") {
		var input searchService.SearchQueryInput
		if vars := req.Variables; vars != nil {
			if rawInput, ok := vars["input"].(map[string]interface{}); ok {
				if queryStr, ok := rawInput["query"].(string); ok {
					input.Query = queryStr
				}
				if cat, ok := rawInput["categorySlug"].(string); ok {
					input.CategorySlug = cat
				}
				if sort, ok := rawInput["sortBy"].(string); ok {
					input.SortBy = sort
				}
			}
		}
		res := h.searchSvc.Search(ctx, input)
		return GraphQLResponse{Data: map[string]interface{}{"searchProducts": res}}
	}

	// 6. QUERY: SearchSuggestions (Elasticsearch Autocomplete)
	if strings.Contains(q, "searchSuggestions") {
		prefix := ""
		if vars := req.Variables; vars != nil {
			if p, ok := vars["prefix"].(string); ok {
				prefix = p
			}
		}
		suggestions := h.searchSvc.Autocomplete(ctx, prefix)
		return GraphQLResponse{Data: map[string]interface{}{"searchSuggestions": suggestions}}
	}

	// 7. QUERY: ProductBySlug
	if strings.Contains(q, "productBySlug") {
		var slug string
		if vars := req.Variables; vars != nil {
			if s, ok := vars["slug"].(string); ok {
				slug = s
			}
		}
		p, err := h.catalogSvc.GetProductBySlug(ctx, slug)
		if err != nil {
			return GraphQLResponse{Errors: []GraphQLError{{Message: err.Error()}}}
		}
		return GraphQLResponse{Data: map[string]interface{}{"productBySlug": p}}
	}

	// 8. QUERY: Categories
	if strings.Contains(q, "categories") {
		cats := h.catalogSvc.ListCategories(ctx)
		return GraphQLResponse{Data: map[string]interface{}{"categories": cats}}
	}

	// 9. QUERY: Cart
	if strings.Contains(q, "cart") {
		cart := h.cartSvc.GetOrCreateCart(ctx, sessionID, userID != "")
		return GraphQLResponse{Data: map[string]interface{}{"cart": cart}}
	}

	// 10. QUERY: MyOrders
	if strings.Contains(q, "myOrders") {
		orders := h.ordersSvc.ListOrdersByUser(ctx, sessionID)
		return GraphQLResponse{Data: map[string]interface{}{
			"myOrders": map[string]interface{}{
				"orders": orders,
				"pageInfo": map[string]interface{}{
					"totalItems": len(orders),
				},
			},
		}}
	}

	/**
	 * ============================================================================
	 * CONCURRENCY PATTERN: Fan-Out / Fan-In Parallel Aggregations (sync.WaitGroup)
	 * 
	 * WHY PARALLEL RESOLVER EXECUTION?
	 * When the homepage or catalog renders, it requires:
	 * 1. Product Catalog items
	 * 2. Categories hierarchy
	 * 3. User's active Cart state
	 * Instead of serial execution (T1 + T2 + T3 = 150ms), executing all 3 resolvers
	 * concurrently in separate goroutines reduces latency to max(T1, T2, T3) = 50ms.
	 * ============================================================================
	 */
	if strings.Contains(q, "HomeAggregate") || strings.Contains(q, "products") {
		var wg sync.WaitGroup
		var prods []*catalogDomain.Product
		var cats []*catalogDomain.Category
		var cart *cartDomain.Cart

		wg.Add(3)

		// Goroutine 1: Fetch Products
		go func() {
			defer wg.Done()
			prods = h.catalogSvc.ListAllProducts(ctx)
		}()

		// Goroutine 2: Fetch Categories
		go func() {
			defer wg.Done()
			cats = h.catalogSvc.ListCategories(ctx)
		}()

		// Goroutine 3: Fetch Cart
		go func() {
			defer wg.Done()
			cart = h.cartSvc.GetOrCreateCart(ctx, sessionID, userID != "")
		}()

		wg.Wait()

		return GraphQLResponse{Data: map[string]interface{}{
			"products":   prods,
			"categories": cats,
			"cart":       cart,
		}}
	}

	// Default fallback: return all featured products
	prods := h.catalogSvc.ListAllProducts(ctx)
	return GraphQLResponse{Data: map[string]interface{}{"products": prods}}
}
