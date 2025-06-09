package app

import (
	"kart-challenge/internal/orders"
	"kart-challenge/internal/products"
	"kart-challenge/internal/promos"

	"github.com/gofiber/fiber/v2"
)

type Handlers struct {
	PromoHandler   *promos.Handler
	ProductHandler *products.Handler
	OrderHandler   *orders.Handler
}

func RegisterAPIRoutes(app *fiber.App, h *Handlers) {
	v1 := app.Group("/api/v1")

	// PromoCode APIs
	v1.Post("/promo_code/validate", h.PromoHandler.ValidatePromoCode)

	// Product API
	v1.Get("/products", h.ProductHandler.GetProducts)
	v1.Get("/products/:id", h.ProductHandler.GetProductByID)

	// Order API
	v1.Post("/orders", h.OrderHandler.CreateOrder)
	v1.Get("/orders/:id", h.OrderHandler.GetOrderByID)

	// --- Admin/Debug Endpoints (Optional for internal/testing) ---
	// These would typically be secured with proper authentication/authorization
	// and likely live under a separate admin API group.
	// v1.Post("/admin/promo_code/reload", h.PromoHandler.ReloadPromoCodes) // Assuming a reload method in handler for admin trigger
	// v1.Get("/debug/promo_codes", h.PromoHandler.GetPromoCodeCountsHandler)
	v1.Get("/debug/orders", h.OrderHandler.GetAllOrders)
}
