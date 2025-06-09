package products

import (
	"kart-challenge/internal/domain" // Corrected import path
)

// Service defines the interface for product business logic.
type Service interface {
	GetAllProducts() []domain.Product
	GetProductByID(id string) (domain.Product, bool)
}

// ProductService implements the Service interface.
type ProductService struct {
	repo ProductRepository
}

// NewService creates a new ProductService.
func NewService(repo ProductRepository) Service {
	return &ProductService{repo: repo}
}

func NewInMemoryProductService() *ProductService {
	s := &ProductService{
		repo: NewInMemoryProductRepository(),
	}
	return s
}

// GetAllProducts returns a list of all available products.
func (s *ProductService) GetAllProducts() []domain.Product {
	return s.repo.GetAll()
}

// GetProductByID returns a product by its ID.
func (s *ProductService) GetProductByID(id string) (domain.Product, bool) {
	return s.repo.GetByID(id)
}
