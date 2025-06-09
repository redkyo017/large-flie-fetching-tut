package promos

import (
	_ "github.com/lib/pq" // PostgreSQL driver
)

type PromoCodeRepository interface {
	GetCount(code string) (int, bool)
	IncrementCount(code string)
	Reset() error
	GetAllCounts() map[string]int
	BulkIncrement(codes map[string]int) error
}
