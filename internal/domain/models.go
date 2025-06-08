package domain

type ValidatePromoteCodeRequest struct {
	PromoteCode string `json:"promote_code" validate:"required"`
}

type ValidatePromoCodeResponse struct {
	PromoCode string `json:"promo_code"`
	Valid     bool   `json:"valid"`
	Message   string `json:"message"`
}

type Product struct {
	ID       string  `json:"id"`
	Name     string  `json:"name`
	Price    float64 `json:"price"`
	Category string  `json:"category"`

	// Description string  `json:"description,omitempty"`
	// Category    string  `json:"category"`
	// ImageUrl    string  `json:"image_url,omitempty"`
	// Available   bool    `json:"available"`
}

type OrderLineItem struct {
	ProductID string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

type Order struct {
	ID       string          `json:"id"`
	Items    []OrderLineItem `json:"items"`
	Total    float64         `json:"total"`
	Products []Product       `json:"products"`

	// TotalPrice  float64         `json:"total_price"`
	// OrderStatus string  `json:"order_status"` // e.g., "pending", "completed", "cancelled"
	// CreatedAt   string  `json:"created_at"`   // ISO 8601 format
	// UpdatedAt   string  `json:"updated_at"`   // ISO 8601 format
	// PromoCode   string  `json:"promo_code,omitempty"`
	// Discount    float64         `json:"discount,omitempty"`
	// FinalPrice  float64         `json:"final_price"`
}

type CreateOrderRequest struct {
	CouponCode string          `json:"coupon_code"`
	Items      []OrderLineItem `json:"items"`
}

type ErrorResponse struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}
