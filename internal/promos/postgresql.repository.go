package promos

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
)

// PostgresPromoCodeRepository implements PromoCodeRepository for PostgreSQL.
type PostgresPromoCodeRepository struct {
	db *sql.DB
}

// NewPostgresPromoCodeRepository creates a new PostgreSQL promo code repository.
// It also ensures the table schema exists.
func NewPostgresPromoCodeRepository(databaseURL string) (PromoCodeRepository, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Set connection pool properties
	db.SetMaxOpenConns(25)                 // Max connections to the database
	db.SetMaxIdleConns(10)                 // Max idle connections in the pool
	db.SetConnMaxLifetime(5 * time.Minute) // Max lifetime of a connection

	// Ping the database to ensure connection is valid
	if err = db.Ping(); err != nil {
		db.Close() // Close if ping fails
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Ensure the promo_codes table exists
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS promo_codes (
		code VARCHAR(10) PRIMARY KEY,
		count INTEGER NOT NULL DEFAULT 0
	);`
	if _, err := db.Exec(createTableSQL); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create promo_codes table: %w", err)
	}

	log.Println("PostgreSQL PromoCodeRepository initialized and table checked.")
	return &PostgresPromoCodeRepository{db: db}, nil
}

// GetCount fetches the count from the database.
func (r *PostgresPromoCodeRepository) GetCount(code string) (int, bool) {
	var count int
	err := r.db.QueryRow("SELECT count FROM promo_codes WHERE code = $1", code).Scan(&count)
	if err == sql.ErrNoRows {
		return 0, false
	}
	if err != nil {
		log.Printf("ERROR: Failed to get promo code count from DB: %v", err)
		return 0, false // Or propagate error if a different error handling strategy is desired
	}
	return count, true
}

// BulkIncrement performs batch upserts to the database.
// This is the core for efficient initial loading of derived data.
func (r *PostgresPromoCodeRepository) BulkIncrement(codes map[string]int) error {
	if len(codes) == 0 {
		return nil // Nothing to update
	}

	// Use a transaction for atomicity and performance
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction for bulk increment: %w", err)
	}
	defer tx.Rollback() // Rollback on error, commit otherwise

	// Prepare the UPSERT statement (INSERT ... ON CONFLICT UPDATE)
	// This is standard for modern PostgreSQL for batch updates.
	// We're passing values dynamically for each row using string building,
	// which is simple but can be inefficient for HUGE numbers of values.
	// For truly massive batches, consider `pgx`'s CopyFrom or `pq.CopyIn` if inserting new.
	// For upserts, `ON CONFLICT` is the way.

	// Dynamically build the VALUES part for bulk insert/upsert
	// Max parameter limit is usually 32767 for PostgreSQL. For huge batches, break into smaller chunks.
	const batchSize = 1000 // Number of (code, count) pairs per batch

	valueStrings := make([]string, 0, batchSize)
	valueArgs := make([]interface{}, 0, batchSize*2) // 2 args per row (code, count)
	argCounter := 0

	var currentBatch int = 0
	for code, increment := range codes {
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d)", argCounter*2+1, argCounter*2+2))
		valueArgs = append(valueArgs, code, increment)
		argCounter++

		if argCounter == batchSize {
			currentBatch++
			log.Printf("Processing batch %d with %d codes...", currentBatch, batchSize)
			stmt := fmt.Sprintf(`
				INSERT INTO promo_codes (code, count)
				VALUES %s
				ON CONFLICT (code) DO UPDATE SET count = promo_codes.count + EXCLUDED.count;
			`, strings.Join(valueStrings, ","))

			_, err := tx.Exec(stmt, valueArgs...)
			if err != nil {
				return fmt.Errorf("failed to execute batch upsert %d: %w", currentBatch, err)
			}

			// Reset for next batch
			valueStrings = make([]string, 0, batchSize)
			valueArgs = make([]interface{}, 0, batchSize*2)
			argCounter = 0
		}
	}

	// Execute any remaining values in the last batch
	if argCounter > 0 {
		currentBatch++
		log.Printf("Processing final batch %d with %d codes...", currentBatch, argCounter)
		stmt := fmt.Sprintf(`
			INSERT INTO promo_codes (code, count)
			VALUES %s
			ON CONFLICT (code) DO UPDATE SET count = promo_codes.count + EXCLUDED.count;
		`, strings.Join(valueStrings, ","))

		_, err := tx.Exec(stmt, valueArgs...)
		if err != nil {
			return fmt.Errorf("failed to execute final batch upsert %d: %w", currentBatch, err)
		}
	}

	return tx.Commit() // Commit the transaction
}

// Reset truncates the promo_codes table.
func (r *PostgresPromoCodeRepository) Reset() error {
	_, err := r.db.Exec("TRUNCATE TABLE promo_codes")
	if err != nil {
		return fmt.Errorf("failed to truncate promo_codes table: %w", err)
	}
	log.Println("promo_codes table truncated.")
	return nil
}

// GetAllCounts fetches all counts from the database.
// WARNING: This can be very slow and memory-intensive for "hundreds of millions" of codes.
// This is primarily for debugging/admin in a large-scale scenario, not regular use.
func (r *PostgresPromoCodeRepository) GetAllCounts() map[string]int {
	counts := make(map[string]int)
	rows, err := r.db.Query("SELECT code, count FROM promo_codes")
	if err != nil {
		log.Printf("ERROR: Failed to get all promo codes from DB: %v", err)
		return counts
	}
	defer rows.Close()

	for rows.Next() {
		var code string
		var count int
		if err := rows.Scan(&code, &count); err != nil {
			log.Printf("ERROR: Failed to scan promo code row: %v", err)
			continue
		}
		counts[code] = count
	}
	log.Printf("Retrieved %d promo codes from database (for GetAllCounts).", len(counts))
	return counts
}

// Close closes the database connection pool.
func (r *PostgresPromoCodeRepository) Close() error {
	if r.db != nil {
		log.Println("Closing PostgreSQL database connection pool...")
		return r.db.Close()
	}
	return nil
}

func (r *PostgresPromoCodeRepository) IncrementCount(code string) {
}
