package products

import (
	"kart-challenge/internal/domain"
	"log"
	"sync"
)

// ProductRepository defines the interface for product data access.
type ProductRepository interface {
	GetAll() []domain.Product
	GetByID(id string) (domain.Product, bool)
}

// inMemoryProductRepository is an in-memory implementation of ProductRepository.
type inMemoryProductRepository struct {
	products map[string]domain.Product
	mu       sync.RWMutex
}

// NewInMemoryProductRepository creates a new in-memory product repository with mock data.
func NewInMemoryProductRepository() ProductRepository {
	r := &inMemoryProductRepository{
		products: make(map[string]domain.Product),
	}
	r.loadMockProducts() // Load some initial mock products
	return r
}

// loadMockProducts populates the repository with some example products.
func (r *inMemoryProductRepository) loadMockProducts() {
	r.mu.Lock()
	defer r.mu.Unlock()

	products := []domain.Product{
		{ID: "prod1", Name: "Burger Classic", Price: 12.99, Category: "Burgers"},
		{ID: "prod2", Name: "Fries Large", Price: 3.49, Category: "Sides"},
		{ID: "prod3", Name: "Coca-Cola", Price: 2.50, Category: "Drinks"},
		{ID: "prod4", Name: "Veggie Burger", Price: 11.50, Category: "Burgers"},
		{ID: "prod5", Name: "Chicken Nuggets", Price: 6.00, Category: "Sides"},
	}

	for _, p := range products {
		r.products[p.ID] = p
	}
	log.Printf("Loaded %d mock products into repository.", len(r.products))
}

// GetAll returns a list of all available products.
func (r *inMemoryProductRepository) GetAll() []domain.Product {
	r.mu.RLock()
	defer r.mu.RUnlock()

	allProducts := make([]domain.Product, 0, len(r.products))
	for _, p := range r.products {
		allProducts = append(allProducts, p)
	}
	return allProducts
}

// GetByID returns a product by its ID.
func (r *inMemoryProductRepository) GetByID(id string) (domain.Product, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	product, found := r.products[id]
	return product, found
}
