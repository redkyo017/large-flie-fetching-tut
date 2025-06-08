package domain

import "errors"

// Define custom errors for common scenarios to allow for better error handling and mapping to HTTP responses.
var (
	ErrProductNotFound         = errors.New("product not found")
	ErrProductUnavailable      = errors.New("product is currently unavailable")
	ErrInvalidQuantity         = errors.New("quantity for product must be positive")
	ErrOrderNotFound           = errors.New("order not found")
	ErrInvalidPromoCode        = errors.New("invalid promo code") // More specific message handled by service
	ErrPromoCodeLength         = errors.New("promo code length invalid")
	ErrPromoCodeNotEnoughFiles = errors.New("promo code not found in enough files")
	ErrInvalidRequestPayload   = errors.New("invalid request payload")
	ErrInternalServerError     = errors.New("internal server error")
)
