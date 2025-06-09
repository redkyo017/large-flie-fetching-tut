package products

import (
	"kart-challenge/internal/domain"
	"testing"
)

// Mock ProductRepository for testing the Service in isolation
type mockProductRepository struct {
	products map[string]domain.Product
}

func (m *mockProductRepository) GetAll() []domain.Product {
	allProducts := make([]domain.Product, 0, len(m.products))
	for _, p := range m.products {
		allProducts = append(allProducts, p)
	}
	return allProducts
}

func (m *mockProductRepository) GetByID(id string) (domain.Product, bool) {
	product, found := m.products[id]
	return product, found
}

func TestProductService_GetAllProducts(t *testing.T) {
	mockProducts := map[string]domain.Product{
		"p1": {ID: "p1", Name: "Test Product 1", Price: 10.0},
		"p2": {ID: "p2", Name: "Test Product 2", Price: 20.0},
	}
	mockRepo := &mockProductRepository{products: mockProducts}
	service := NewService(mockRepo)

	products := service.GetAllProducts()

	if len(products) != 2 {
		t.Errorf("Expected 2 products, got %d", len(products))
	}
	// Check if the expected products are present (order doesn't matter for slice from map)
	foundP1, foundP2 := false, false
	for _, p := range products {
		if p.ID == "p1" && p.Name == "Test Product 1" {
			foundP1 = true
		}
		if p.ID == "p2" && p.Name == "Test Product 2" {
			foundP2 = true
		}
	}
	if !foundP1 || !foundP2 {
		t.Errorf("Did not find all expected products in the returned list")
	}
}

func TestProductService_GetProductByID(t *testing.T) {
	mockProducts := map[string]domain.Product{
		"p1": {ID: "p1", Name: "Test Product 1", Price: 10.0},
		"p2": {ID: "p2", Name: "Test Product 2", Price: 20.0},
	}
	mockRepo := &mockProductRepository{products: mockProducts}
	service := NewService(mockRepo)

	tests := []struct {
		id            string
		expectedFound bool
		expectedName  string
	}{
		{"p1", true, "Test Product 1"},
		{"p2", true, "Test Product 2"},
		{"p3", false, ""}, // Not found
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			product, found := service.GetProductByID(tt.id)
			if found != tt.expectedFound {
				t.Errorf("Expected found status %t, got %t for ID %s", tt.expectedFound, found, tt.id)
			}
			if found && product.Name != tt.expectedName {
				t.Errorf("Expected product name %s, got %s for ID %s", tt.expectedName, product.Name, tt.id)
			}
		})
	}
}
