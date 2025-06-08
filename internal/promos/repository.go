package promos

import "sync"

type PromoCodeRepository interface {
	GetCount(code string) (int, bool)
	IncreamentCount(code string)
	Reset()
	GetAllCounts() map[string]int
}

type inMemoryPromoCodeRepository struct {
	promoCodeCounts map[string]int
	mu              sync.RWMutex
}

func NewInMemoryPromoCodeRepository() *inMemoryPromoCodeRepository {
	return &inMemoryPromoCodeRepository{
		promoCodeCounts: make(map[string]int),
	}
}

func (r *inMemoryPromoCodeRepository) GetCount(code string) (int, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	count, exists := r.promoCodeCounts[code]
	return count, exists
}

func (r *inMemoryPromoCodeRepository) IncreamentCount(code string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.promoCodeCounts[code]++
}

func (r *inMemoryPromoCodeRepository) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.promoCodeCounts = make(map[string]int)
}

func (r *inMemoryPromoCodeRepository) GetAllCounts() map[string]int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	copied := make(map[string]int, len(r.promoCodeCounts))
	for code, count := range r.promoCodeCounts {
		copied[code] = count
	}

	return copied
}
