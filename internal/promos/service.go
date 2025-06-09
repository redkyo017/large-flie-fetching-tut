package promos

import (
	"bufio"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/allegro/bigcache/v3"
)

const (
	promoCodeMinLength = 8
	promoCodeMaxLength = 10
	bigCacheName       = "promoCodeValidationCache"

	aggregationBatchSize = 100000 // Process 100,000 unique codes at a time for aggregation
)

type Service interface {
	LoadPromoCodesFromURLs(urls []string) error
	ValidatePromoCode(code string) (bool, string)
	GetPromoCodeCounts() map[string]int
	Close() error //closing resources like BigCache
}

type PromoCodeService struct {
	repo                    PromoCodeRepository
	bigCache                *bigcache.BigCache
	maxDecompressedFileSize int64
	environment             string // "development" or "production"
	localCouponDirPath      string // Path to local .gz coupon files
}

func NewService(maxDecompressedFileSizeMB int, env string, localCouponDirPath string) Service {
	cacheConfig := bigcache.DefaultConfig(1 * time.Hour)
	cacheConfig.CleanWindow = 10 * time.Minute
	cacheConfig.Verbose = true
	cacheConfig.OnRemoveWithReason = func(key string, entry []byte, reason bigcache.RemoveReason) {
		// log.Printf("BigCache entry '%s' removed due to: %s", key, reason.String())
		log.Printf("BigCache entry '%s' removed due to: %s", key, fmt.Sprintf("%v", reason))
	}

	bc, err := bigcache.New(context.Background(), cacheConfig)
	if err != nil {
		log.Fatalf("Failed to initialize BigCache for %s: %v", bigCacheName, err)
	}

	return &PromoCodeService{
		repo:                    NewInMemoryPromoCodeRepository(), // Use the in-memory repository
		bigCache:                bc,
		maxDecompressedFileSize: int64(maxDecompressedFileSizeMB) * 1024 * 1024,
		environment:             env,
		localCouponDirPath:      localCouponDirPath,
	}
}

func (s *PromoCodeService) LoadPromoCodesFromURLs(urls []string) error {
	log.Println("Starting to load promo codes from URLs...")

	s.repo.Reset()
	var wg sync.WaitGroup

	codesCh := make(chan map[string]bool, len(urls))
	errCh := make(chan error, len(urls))

	for i, url := range urls {
		wg.Add(1)
		go func(fileIndex int, u string) {
			defer wg.Done()
			log.Printf("Processing file %d: %s", fileIndex+1, u)
			fileFoundCodes, err := s.processSinglePromoFile(fileIndex, u)
			if err != nil {
				errCh <- fmt.Errorf("error processing file %d (%s): %w", fileIndex+1, u, err)
			}
			codesCh <- fileFoundCodes
		}(i, url)
	}

	// Wait for all goroutines to complete and close channels
	go func() {
		wg.Wait()
		close(codesCh)
		close(errCh)
	}()

	// Collect all fileFoundCodes maps from the channel
	// // This runs concurrently with the file processing but doesn't touch the repo's mutex yet.
	// allCollectedCodes := make([]map[string]bool, 0, len(urls))
	// for fileFoundCodes := range codesCh {
	// 	allCollectedCodes = append(allCollectedCodes, fileFoundCodes)
	// }

	// // --- Final Aggregation into the Repository (Single Locked Operation) ---
	// // After all concurrent file processing is done and results collected,
	// // perform a single bulk update to the repository.
	// log.Println("Aggregating results into main promo code repository...")
	// s.repo.(*inMemoryPromoCodeRepository).mu.Lock() // Access the underlying mutex directly for this optimized bulk op
	// for _, fileFoundCodes := range allCollectedCodes {
	// 	for code := range fileFoundCodes {
	// 		s.repo.(*inMemoryPromoCodeRepository).promoCodeCounts[code]++ // Directly increment
	// 	}
	// }
	// s.repo.(*inMemoryPromoCodeRepository).mu.Unlock()
	// log.Println("Aggregation complete.")
	log.Println("Starting batched aggregation into promo code repository...")
	currentBatch := make(map[string]int)
	processedCount := 0

	for fileFoundCodes := range codesCh { // This loop runs after individual files are processed concurrently
		for code := range fileFoundCodes {
			currentBatch[code]++
			processedCount++

			// If current batch size reaches the limit, perform bulk increment
			if len(currentBatch) >= aggregationBatchSize {
				log.Printf("Aggregating %d unique codes into repository (processed so far: %d)...", len(currentBatch), processedCount)
				if err := s.repo.BulkIncrement(currentBatch); err != nil {
					return fmt.Errorf("failed to perform bulk increment on repository: %w", err)
				}
				currentBatch = make(map[string]int) // Reset batch
			}
		}
	}

	// Perform final bulk increment for any remaining codes in the batch
	if len(currentBatch) > 0 {
		log.Printf("Performing final aggregation of %d unique codes into repository (total processed: %d)...", len(currentBatch), processedCount)
		if err := s.repo.BulkIncrement(currentBatch); err != nil {
			return fmt.Errorf("failed to perform final bulk increment on repository: %w", err)
		}
	}
	log.Println("Batched aggregation complete.")

	var allErrors error
	for err := range errCh {
		if allErrors == nil {
			allErrors = err
		} else {
			allErrors = fmt.Errorf("%v; %w", allErrors, err)
		}
	}
	log.Printf("Finished loading promo codes. Total unique codes found: %d", len(s.repo.GetAllCounts()))
	return allErrors
}

func (s *PromoCodeService) processSinglePromoFile(fileIndex int, url string) (map[string]bool, error) {
	var reader io.ReadCloser
	var source string

	// Determine if we should load from local disk or remote URL
	if s.environment == "development" {
		fileName := filepath.Base(url) // e.g., "couponbase1.gz"
		localPath := filepath.Join(s.localCouponDirPath, fileName)

		if _, err := os.Stat(localPath); err == nil {
			log.Printf("INFO: Loading file %d from local path: %s", fileIndex+1, localPath)
			file, err := os.Open(localPath)
			if err != nil {
				return nil, fmt.Errorf("failed to open local GZ file '%s': %w", localPath, err)
			}
			reader = file
			source = "local"
		} else {
			log.Printf("WARN: Local file '%s' not found for file %d. Attempting remote download.", localPath, fileIndex+1)
			goto remoteDownload // Jump to remote download if local file not found
		}
	} else {
		// Not development environment, always download remotely
		goto remoteDownload
	}

	// Label for jumping to remote download logic
remoteDownload:
	if reader == nil { // need to download remotely
		log.Printf("INFO: Downloading file %d from remote URL: %s", fileIndex+1, url)
		// Increased timeout for potentially large files and network conditions
		client := &http.Client{Timeout: 5 * time.Minute}
		resp, err := client.Get(url)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch GZ file: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close() // Close body on error
			return nil, fmt.Errorf("remote server returned status code %d", resp.StatusCode)
		}
		reader = resp.Body
		source = "remote"
	}
	defer reader.Close() // Ensure the source reader (local file or http.Response.Body) is closed

	tempFile, err := os.CreateTemp("", fmt.Sprintf("couponbase%d-*.tmp", fileIndex+1))
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary file: %w", err)
	}

	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	gzipReader, err := gzip.NewReader(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzipReader.Close()

	// limitedReader to prevent writing excessively large decompressed files to disk
	limitedReader := io.LimitReader(gzipReader, s.maxDecompressedFileSize)

	log.Printf("Starting decompression from %s source for file %d (%s)...", source, fileIndex+1, url)
	bytesWritten, err := io.Copy(tempFile, limitedReader)

	if err != nil {
		return nil, fmt.Errorf("failed to decompress to temporary file: %w", err)
	}
	tempFile.Close()
	log.Printf("Decompression finished for file %d (%s) from %s. Bytes written: %d", fileIndex+1, url, source, bytesWritten)

	if bytesWritten >= s.maxDecompressedFileSize {
		return nil, fmt.Errorf("decompressed file size exceeded limit of %dMB, possible malicious file", s.maxDecompressedFileSize/1024/1024)
	}
	fileForScan, err := os.Open(tempFile.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to open temporary decompressed file: %w", err)
	}
	defer fileForScan.Close()

	scanner := bufio.NewScanner(fileForScan)
	fileFoundCodes := make(map[string]bool)

	// promoCodeRegex := regexp.MustCompile(`[A-Za-z0-9]{8,10}`)

	for scanner.Scan() {
		line := scanner.Text()
		// matches := promoCodeRegex.FindAllString(line, -1)
		// for _, code := range matches {
		// 	if len(code) >= promoCodeMinLength && len(code) <= promoCodeMaxLength {
		// 		fileFoundCodes[code] = true
		// 	}
		// }
		if len(line) >= promoCodeMinLength && len(line) <= promoCodeMaxLength {
			fileFoundCodes[line] = true
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading decompressed file: %w", err)
	}

	// for code := range fileFoundCodes {
	// 	s.repo.IncrementCount(code)
	// }

	return fileFoundCodes, nil
}

func (s *PromoCodeService) ValidatePromoCode(code string) (bool, string) {
	if len(code) < promoCodeMinLength || len(code) > promoCodeMaxLength {
		return false, fmt.Sprintf("Promo code must be between %d and %d characters long.", promoCodeMinLength, promoCodeMaxLength)
	}
	count, exists := s.repo.GetCount(code)
	if !exists || count < 2 {
		return false, "Promo code not found in at least two files."
	}

	return true, "Promo code is valid."
}
func (s *PromoCodeService) GetPromoCodeCounts() map[string]int {
	return s.repo.GetAllCounts()
}

func (s *PromoCodeService) Close() error {
	log.Println("Closing PromoCodeService BigCache...")
	return s.bigCache.Close()
}
