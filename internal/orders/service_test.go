package orders

import (
	"errors"
	"kart-challenge/internal/domain"
	"testing"
)

// Mock OrderRepository for testing OrderService
type mockOrderRepository struct {
	orders map[string]domain.Order
}

func (m *mockOrderRepository) Create(order domain.Order) error {
	m.orders[order.ID] = order
	return nil
}

func (m *mockOrderRepository) GetByID(id string) (domain.Order, bool) {
	order, found := m.orders[id]
	return order, found
}

func (m *mockOrderRepository) GetAll() []domain.Order {
	allOrders := make([]domain.Order, 0, len(m.orders))
	for _, o := range m.orders {
		allOrders = append(allOrders, o)
	}
	return allOrders
}

// Mock ProductService for OrderService tests
type mockProductService struct {
	products map[string]domain.Product
}

func (m *mockProductService) GetAllProducts() []domain.Product {
	return nil // Not needed for these tests
}

func (m *mockProductService) GetProductByID(id string) (domain.Product, bool) {
	p, found := m.products[id]
	return p, found
}

// Mock PromoCodeService for OrderService tests
type mockPromoCodeService struct {
	validPromoCodes map[string]bool
}

func (m *mockPromoCodeService) LoadPromoCodesFromURLs(urls []string) error {
	return nil // Not needed for these tests
}

func (m *mockPromoCodeService) ValidatePromoCode(code string) (bool, string) {
	isValid := m.validPromoCodes[code]
	if isValid {
		return true, "Promo code is valid."
	}
	return false, "Promo code is invalid."
}

func (m *mockPromoCodeService) GetPromoCodeCounts() map[string]int {
	return nil // Not needed for these tests
}

func (m *mockPromoCodeService) Close() error {
	return nil // No-op
}

func TestOrderService_CreateOrder(t *testing.T) {
	// Setup mocks
	mockProducts := map[string]domain.Product{
		"prod1": {ID: "prod1", Name: "Burger", Price: 10.0},
		"prod2": {ID: "prod2", Name: "Fries", Price: 5.0},
		"prod3": {ID: "prod3", Name: "Unavailable Drink", Price: 3.0},
	}
	productService := &mockProductService{products: mockProducts}

	validPromoCodes := map[string]bool{"PROMO10": true, "PROMO20": true}
	promoCodeService := &mockPromoCodeService{validPromoCodes: validPromoCodes}

	orderRepo := &mockOrderRepository{orders: make(map[string]domain.Order)}
	service := NewService(orderRepo, productService, promoCodeService)

	tests := []struct {
		name             string
		request          domain.CreateOrderRequest
		expectedErr      error
		expectedPrice    float64
		expectedDiscount float64
	}{
		{
			name: "Successful order with no promo code",
			request: domain.CreateOrderRequest{
				Items: []domain.OrderLineItem{
					{ProductID: "prod1", Quantity: 1},
					{ProductID: "prod2", Quantity: 2},
				},
			},
			expectedErr: nil,
		},
		{
			name: "Successful order with valid promo code (PROMO10)",
			request: domain.CreateOrderRequest{
				Items: []domain.OrderLineItem{
					{ProductID: "prod1", Quantity: 2},
				},
			},
			expectedErr: nil,
		},
		{
			name: "Order with invalid product ID",
			request: domain.CreateOrderRequest{
				Items: []domain.OrderLineItem{
					{ProductID: "nonexistent", Quantity: 1},
				},
			},
			expectedErr: domain.ErrProductNotFound,
		},
		{
			name: "Order with zero quantity",
			request: domain.CreateOrderRequest{
				Items: []domain.OrderLineItem{
					{ProductID: "prod1", Quantity: 0},
				},
			},
			expectedErr: domain.ErrInvalidQuantity,
		},
		{
			name: "Order with negative quantity",
			request: domain.CreateOrderRequest{
				Items: []domain.OrderLineItem{
					{ProductID: "prod1", Quantity: -1},
				},
			},
			expectedErr: domain.ErrInvalidQuantity,
		},
		{
			name: "Order with invalid promo code",
			request: domain.CreateOrderRequest{
				Items: []domain.OrderLineItem{
					{ProductID: "prod1", Quantity: 1},
				},
				CouponCode: "INVALIDCODE",
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			order, err := service.CreateOrder(tt.request)

			if !errors.Is(err, tt.expectedErr) {
				t.Errorf("Expected error %v, got %v", tt.expectedErr, err)
			}

			if tt.expectedErr == nil { // Only check order details if no error was expected
				if order.ID == "" {
					t.Errorf("Expected order ID to be generated, got empty")
				}
				if len(order.Items) != len(tt.request.Items) {
					t.Errorf("Expected %d order items, got %d", len(tt.request.Items), len(order.Items))
				}
			}
		})
	}
}

func TestOrderService_GetOrder(t *testing.T) {
	// Setup mock repository with a pre-existing order
	mockOrder := domain.Order{
		ID: "order123",
	}
	orderRepo := &mockOrderRepository{orders: map[string]domain.Order{"order123": mockOrder}}

	// Mocks for dependencies, not directly used in GetOrder but required by NewService
	productService := &mockProductService{}
	promoCodeService := &mockPromoCodeService{}

	service := NewService(orderRepo, productService, promoCodeService)

	tests := []struct {
		name          string
		orderID       string
		expectedFound bool
		expectedOrder domain.Order
	}{
		{
			name:          "Found existing order",
			orderID:       "order123",
			expectedFound: true,
			expectedOrder: mockOrder,
		},
		{
			name:          "Order not found",
			orderID:       "nonexistent",
			expectedFound: false,
			expectedOrder: domain.Order{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			order, found := service.GetOrder(tt.orderID)
			if found != tt.expectedFound {
				t.Errorf("Expected found status %t, got %t", tt.expectedFound, found)
			}
			if found && order.ID != tt.expectedOrder.ID { // Simple ID check for equality
				t.Errorf("Expected order ID %s, got %s", tt.expectedOrder.ID, order.ID)
			}
		})
	}
}

func TestOrderService_GetAllOrders(t *testing.T) {
	// Setup mock repository with pre-existing orders
	mockOrders := map[string]domain.Order{
		"order1": {ID: "order1"},
		"order2": {ID: "order2"},
	}
	orderRepo := &mockOrderRepository{orders: mockOrders}

	// Mocks for dependencies, not directly used in GetAllOrders but required by NewService
	productService := &mockProductService{}
	promoCodeService := &mockPromoCodeService{}

	service := NewService(orderRepo, productService, promoCodeService)

	orders := service.GetAllOrders()

	if len(orders) != len(mockOrders) {
		t.Errorf("Expected %d orders, got %d", len(mockOrders), len(orders))
	}

	// Check if all mock orders are present in the returned slice
	foundCount := 0
	for _, o := range orders {
		if _, exists := mockOrders[o.ID]; exists {
			foundCount++
		}
	}
	if foundCount != len(mockOrders) {
		t.Errorf("Not all expected orders were found in the returned list")
	}
}
