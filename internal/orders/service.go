package orders

import (
	"fmt"
	"log"
	"sync"

	"kart-challenge/internal/domain" // Corrected import path
	"kart-challenge/internal/products"
	"kart-challenge/internal/promos"

	"github.com/google/uuid"
)

// Service defines the interface for order business logic.
type Service interface {
	CreateOrder(req domain.CreateOrderRequest) (domain.Order, error)
	GetOrder(orderID string) (domain.Order, bool)
	GetAllOrders() []domain.Order
}

// OrderService implements the Service interface.
type OrderService struct {
	repo             OrderRepository
	mu               sync.RWMutex
	ProductService   products.Service // Dependency to get product details
	PromoCodeService promos.Service   // Dependency to validate promo codes
}

// NewService creates a new OrderService.
func NewService(repo OrderRepository, productService products.Service, promoCodeService promos.Service) Service {
	return &OrderService{
		repo:             repo,
		ProductService:   productService,
		PromoCodeService: promoCodeService,
	}
}

// CreateOrder processes a new order request.
func (s *OrderService) CreateOrder(req domain.CreateOrderRequest) (domain.Order, error) {
	newOrder := domain.Order{
		ID:    uuid.New().String(), // Generate a unique UUID for the order
		Items: make([]domain.OrderLineItem, 0, len(req.Items)),
		// OrderStatus: "pending",
		// CreatedAt:   time.Now().Format(time.RFC3339), // ISO 8601 format
		// UpdatedAt:   time.Now().Format(time.RFC3339),
	}

	// var totalPrice float64 = 0.0
	// Validate and calculate total price for line items
	for _, itemReq := range req.Items {
		product, found := s.ProductService.GetProductByID(itemReq.ProductID)
		if !found {
			return domain.Order{}, domain.ErrProductNotFound // Use custom error
		}
		if itemReq.Quantity <= 0 {
			return domain.Order{}, domain.ErrInvalidQuantity
		}

		lineItem := domain.OrderLineItem{
			ProductID: itemReq.ProductID,
			Quantity:  itemReq.Quantity,
		}
		newOrder.Items = append(newOrder.Items, lineItem)
		newOrder.Products = append(newOrder.Products, product)
		// totalPrice += product.Price * float64(itemReq.Quantity)
	}
	// newOrder.TotalPrice = totalPrice
	// newOrder.FinalPrice = totalPrice // Initialize final price before discount

	// Apply promo code if provided
	if req.CouponCode != "" {
		isValid, _ := s.PromoCodeService.ValidatePromoCode(req.CouponCode)
		if isValid {
			// newOrder.PromoCode = req.CouponCode
			// Apply a simple discount (e.g., 10% off for valid promo codes)
			// discountAmount := newOrder.TotalPrice * 0.10
			// newOrder.Discount = discountAmount
			// newOrder.FinalPrice = newOrder.TotalPrice - discountAmount
			// log.Printf("Order %s: Promo code '%s' applied. Discount: %.2f, Final Price: %.2f", newOrder.ID, req.PromoCode, newOrder.Discount, newOrder.FinalPrice)
			log.Printf("Order %s: Promo code '%s' applied", newOrder.ID, req.CouponCode)
		} else {
			// Even if promo code is invalid, we proceed with the order without discount
			log.Printf("Order %s: Invalid promo code '%s' provided. Proceeding without discount.", newOrder.ID, req.CouponCode)
			// Optionally, you might want to return an error here or add a specific message to the order object
			// newOrder.Message = "Invalid promo code, order created without discount."
		}
	}

	if err := s.repo.Create(newOrder); err != nil {
		return domain.Order{}, fmt.Errorf("failed to save order: %w", err)
	}

	log.Printf("Created new order: %s", newOrder.ID)
	return newOrder, nil
}

// GetOrder implements the logic to retrieve an order by its ID.
func (s *OrderService) GetOrder(orderID string) (domain.Order, bool) {
	return s.repo.GetByID(orderID)
}

// GetAllOrders implements the logic to retrieve all orders (for admin/debug purposes).
func (s *OrderService) GetAllOrders() []domain.Order {
	return s.repo.GetAll()
}
