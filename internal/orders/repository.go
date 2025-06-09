package orders

import (
	"fmt"
	"kart-challenge/internal/domain"
	"sync"
)

// OrderRepository defines the interface for order data access.
type OrderRepository interface {
	Create(order domain.Order) error
	GetByID(id string) (domain.Order, bool)
	GetAll() []domain.Order
}

// inMemoryOrderRepository is an in-memory implementation of OrderRepository.
type inMemoryOrderRepository struct {
	orders map[string]domain.Order
	mu     sync.RWMutex
}

// NewInMemoryOrderRepository creates a new in-memory order repository.
func NewInMemoryOrderRepository() OrderRepository {
	return &inMemoryOrderRepository{
		orders: make(map[string]domain.Order),
	}
}

// Create adds a new order to the repository.
func (r *inMemoryOrderRepository) Create(order domain.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.orders[order.ID]; exists {
		return fmt.Errorf("order with ID %s already exists", order.ID) // Should not happen with UUID
	}
	r.orders[order.ID] = order
	return nil
}

// GetByID retrieves an order by its ID.
func (r *inMemoryOrderRepository) GetByID(id string) (domain.Order, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	order, found := r.orders[id]
	return order, found
}

// GetAll retrieves all orders.
func (r *inMemoryOrderRepository) GetAll() []domain.Order {
	r.mu.RLock()
	defer r.mu.RUnlock()
	allOrders := make([]domain.Order, 0, len(r.orders))
	for _, order := range r.orders {
		allOrders = append(allOrders, order)
	}
	return allOrders
}
