package products

import (
	"kart-challenge/internal/domain" // Corrected import path

	"github.com/gofiber/fiber/v2"
)

// Handler manages HTTP requests for products.
type Handler struct {
	Service Service
}

// NewHandler creates a new ProductHandler.
func NewHandler(service Service) *Handler {
	return &Handler{Service: service}
}

// GetProducts handles GET /products to retrieve all products.
func (h *Handler) GetProducts(c *fiber.Ctx) error {
	products := h.Service.GetAllProducts()
	return c.Status(fiber.StatusOK).JSON(products)
}

// GetProductByID handles GET /products/:id to retrieve a single product by ID.
func (h *Handler) GetProductByID(c *fiber.Ctx) error {
	id := c.Params("id")
	product, found := h.Service.GetProductByID(id)
	if !found {
		return c.Status(fiber.StatusNotFound).JSON(domain.ErrorResponse{
			Message: domain.ErrProductNotFound.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(product)
}
