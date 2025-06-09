package promos

import (
	"kart-challenge/internal/domain"
	"log"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	Service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{
		Service: service,
	}
}

func (h *Handler) ValidatePromoCode(c *fiber.Ctx) error {
	req := new(domain.ValidatePromoteCodeRequest)

	if err := c.BodyParser(req); err != nil {
		log.Printf("Error parsing promo code validation request: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Message: domain.ErrInvalidRequestPayload.Error() + ": Ensure 'promo_code' is a string.",
		})
	}

	if req.PromoteCode == "" {
		c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Message: "Promo code cannot be empty",
		})
	}

	isValid, message := h.Service.ValidatePromoCode((req.PromoteCode))

	return c.Status(fiber.StatusOK).JSON(domain.ValidatePromoCodeResponse{
		Valid:     isValid,
		Message:   message,
		PromoCode: req.PromoteCode,
	})
}

func (h *Handler) GetPromoCodeCountsHandlers(c *fiber.Ctx) error {
	counts := h.Service.GetPromoCodeCounts()

	return c.Status(fiber.StatusOK).JSON(counts)
}
