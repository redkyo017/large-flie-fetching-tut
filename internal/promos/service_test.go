package promos

import (
	"testing"
)

func TestPromoCodeService_ValidatePromoCode(t *testing.T) {
	service := NewService(1, "production", "") // Use small size, production env (no local files needed for validate)
	defer service.Close()

	// Manually set counts for testing validation logic
	inMemRepo := service.(*PromoCodeService).repo.(*inMemoryPromoCodeRepository)
	inMemRepo.mu.Lock()
	inMemRepo.promoCodeCounts["VALIDCODE"] = 2
	inMemRepo.promoCodeCounts["SINGLEFILE"] = 1
	inMemRepo.promoCodeCounts["LONGCODEEXAMPLE"] = 15 // Too long
	inMemRepo.promoCodeCounts["SHORT"] = 5            // Too short
	inMemRepo.mu.Unlock()

	tests := []struct {
		code          string
		expectedValid bool
		expectedMsg   string
	}{
		{"VALIDCODE", true, "Promo code is valid."},
		{"SINGLEFILE", false, "Promo code not found in at least two files."},
		{"TOOLONGCODE", false, "Promo code must be between 8 and 10 characters long."},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			isValid, msg := service.ValidatePromoCode(tt.code)
			if isValid != tt.expectedValid {
				t.Errorf("For code '%s', expected valid %t, got %t", tt.code, tt.expectedValid, isValid)
			}
			if msg != tt.expectedMsg {
				t.Errorf("For code '%s', expected message '%s', got '%s'", tt.code, tt.expectedMsg, msg)
			}
		})
	}
}
