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
	"regexp"
	"sync"
	"time"
)

const (
	promoCodeMinLength = 8
	promoCodeMaxLength = 10
	bigCacheName       = "promoCodeValidationCache"
)

type Service interface {
	LoadPromoCodesFromURLs(urls []string) error
	ValidatePromoCode(code string) (bool, string)
	GetPromoCodeCounts() map[string]int
	Close() error //closing resources like BigCache
}

type PromoCodeService struct {
	repo                    PromoCodeRepository
	bigCache                *bigCache.BigCache
	maxDecompressedFileSize int64
}

func NewService(maxDecompressedFileSizeMB int) Service {
	cacheConfig := bigcache.DefaultConfig(10 * time.Hour)
	cacheConfig.CleanWind = 10 * time.Minute
	cacheConfig.Verbose = true
	cacheConfig.OnRemoveWithReason = func(key string, entry []byte, reason bigcache.RemoveReason) {
		log.Printf("BigCache entry '%s' removed due to: %s", key, reason.String())
	}

	bc, err := bigcache.New(context.Background(), cacheConfig)
	if err != nil {
		log.Fatalf("Failed to initialize BigCache for %s: %v", bigCacheName, err)
	}

	return &PromoCodeService{
		repo:                    NewInMemoryPromoCodeRepository(), // Use the in-memory repository
		bigCache:                bc,
		maxDecompressedFileSize: int64(maxDecompressedFileSizeMB) * 1024 * 1024,
	}
}

func (s *PromoCodeService) LoadPromoCodesFromURLs(urls []string) error {
	log.Println("Starting to load promo codes from URLs...")

	s.repo.Reset()
	var wg sync.WaitGroup
	errCh := make(chan error, len(urls))

	for i, url := range urls {
		wg.Add(1)
		go func(fileIndex int, u string) {
			defer wg.Done()
			log.Printf("Processing file %d: %s", fileIndex+1, u)
			if err := s.processSinglePromoFile(fileIndex, u); err != nil {
				errCh <- fmt.Errorf("error processing file %d (%s): %w", fileIndex+1, u, err)
			}
		}(i, url)
	}
	wg.Wait()
	close(errCh)

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

func (s *PromoCodeService) processSinglePromoFile(fileIndex int, url string) error {
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch file: %w", err)
	}
	defer resp.Body.Close()

	tempFile, err := os.CreateTemp("", fmt.Sprintf("couponbase%d-*.tmp", fileIndex+1))
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}

	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	gzipReader, err := gzip.NewReader(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzipReader.Close()

	// limitedReader to prevent writing excessively large decompressed files to disk
	limitedReader := io.LimitReader(gzipReader, s.maxDecompressedFileSize)

	bytesWritten, err := io.Copy(tempFile, limitedReader)
	if err != nil {
		return fmt.Errorf("failed to decompress to temporary file: %w", err)
	}
	tempFile.Close()

	if bytesWritten >= s.maxDecompressedFileSize {
		return fmt.Errorf("decompressed file size exceeded limit of %dMB, possible malicious file", s.maxDecompressedFileSize/1024/1024)
	}

	fileForScan, err := os.Open(tempFile.Name())
	if err != nil {
		return fmt.Errorf("failed to open temporary decompressed file: %w", err)
	}
	defer fileForScan.Close()

	scanner := bufio.NewScanner(fileForScan)
	fileFoundCodes := make(map[string]bool)

	promoCodeRegex := regexp.MustCompile(`[A-Za-z0-9]{8,10}`)

	for scanner.Scan() {
		line := scanner.Text()
		matches := promoCodeRegex.FindAllString(line, -1)
		for _, code := range matches {
			if len(code) >= promoCodeMinLength && len(code) <= promoCodeMaxLength {
				fileFoundCodes[code] = true
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading decompressed file: %w", err)
	}

	for code := range fileFoundCodes {
		s.repo.IncreamentCount(code)
	}

	return nil
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
