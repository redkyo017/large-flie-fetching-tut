package orders

import (
	"errors"
	"kart-challenge/internal/domain" // Corrected import path
	"log"

	"github.com/gofiber/fiber/v2"
)

// Handler manages HTTP requests for orders.
type Handler struct {
	Service Service
}

// NewHandler creates a new OrderHandler.
func NewHandler(service Service) *Handler {
	return &Handler{Service: service}
}

// CreateOrder handles POST /orders to create a new order.
func (h *Handler) CreateOrder(c *fiber.Ctx) error {
	req := new(domain.CreateOrderRequest)
	if err := c.BodyParser(req); err != nil {
		log.Printf("Error parsing create order request: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Message: domain.ErrInvalidRequestPayload.Error(),
		})
	}

	// Basic validation
	// if req.CustomerID == "" || len(req.Items) == 0 {
	// 	return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
	// 		Message: "Customer ID and at least one item are required to create an order.",
	// 	})
	// }

	for _, item := range req.Items {
		if item.ProductID == "" || item.Quantity <= 0 {
			return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
				Message: "All order items must have a product ID and a positive quantity.",
			})
		}
	}

	order, err := h.Service.CreateOrder(*req)
	if err != nil {
		log.Printf("Error creating order: %v", err)
		var errMsg string
		var statusCode int

		switch {
		case errors.Is(err, domain.ErrProductNotFound):
			errMsg = domain.ErrProductNotFound.Error()
			statusCode = fiber.StatusNotFound
		case errors.Is(err, domain.ErrProductUnavailable):
			errMsg = domain.ErrProductUnavailable.Error()
			statusCode = fiber.StatusConflict // 409 Conflict for resource state
		case errors.Is(err, domain.ErrInvalidQuantity):
			errMsg = domain.ErrInvalidQuantity.Error()
			statusCode = fiber.StatusBadRequest
		default:
			errMsg = domain.ErrInternalServerError.Error()
			statusCode = fiber.StatusInternalServerError
		}

		return c.Status(statusCode).JSON(domain.ErrorResponse{
			Message: errMsg,
		})
	}

	return c.Status(fiber.StatusCreated).JSON(order)
}

// GetOrderByID handles GET /orders/:id to retrieve a single order by ID.
func (h *Handler) GetOrderByID(c *fiber.Ctx) error {
	id := c.Params("id")
	order, found := h.Service.GetOrder(id)
	if !found {
		return c.Status(fiber.StatusNotFound).JSON(domain.ErrorResponse{
			Message: domain.ErrOrderNotFound.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(order)
}

// GetAllOrders (optional for admin/debug)
func (h *Handler) GetAllOrders(c *fiber.Ctx) error {
	orders := h.Service.GetAllOrders()
	return c.Status(fiber.StatusOK).JSON(orders)
}
