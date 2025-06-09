package promos

import (
	"sync"

	_ "github.com/lib/pq" // PostgreSQL driver
)

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

func (r *inMemoryPromoCodeRepository) IncrementCount(code string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.promoCodeCounts[code]++
}

func (r *inMemoryPromoCodeRepository) Reset() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.promoCodeCounts = make(map[string]int)
	return nil
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

// BulkIncrement increments counts for multiple promo codes atomically.
// It acquires a single write lock for the entire batch.
func (r *inMemoryPromoCodeRepository) BulkIncrement(codes map[string]int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for code, count := range codes {
		r.promoCodeCounts[code] += count
	}
	return nil
}
