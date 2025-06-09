package config

import (
	"log"
	"os"
	"strconv"
)

type Appconfig struct {
	Port               string
	Environment        string //"development", "production"
	CouponFileURLs     []string
	LocalCouponDirPath string // Path to local .gz coupon files (e.g., "./local_coupons")
	MaxFileSizeMB      int
}

func LoadConfig() *Appconfig {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	env := os.Getenv("APP_ENV")
	if env == "" {
		// env = "production" // Default to production if not set
		env = "development" // Default to production if not set
	}
	log.Printf("INFO: Application environment set to '%s'", env)

	couponURLs := []string{
		"https://orderfoodonline-files.s3.ap-southeast-2.amazonaws.com/couponbase1.gz",
		"https://orderfoodonline-files.s3.ap-southeast-2.amazonaws.com/couponbase2.gz",
		"https://orderfoodonline-files.s3.ap-southeast-2.amazonaws.com/couponbase3.gz",
	}

	// Default local directory for coupon files
	localCouponDirPath := os.Getenv("LOCAL_COUPON_DIR")
	if localCouponDirPath == "" {
		localCouponDirPath = "./local_coupons" // Default path relative to project root
	}
	log.Printf("INFO: Local coupon directory path set to '%s'", localCouponDirPath)

	maxFileSizeMBStr := os.Getenv("MAX_FILE_SIZE_MB")
	maxFileSizeMB, err := strconv.Atoi(maxFileSizeMBStr)
	if err != nil || maxFileSizeMB <= 0 {
		maxFileSizeMB = 2048
		// maxFileSizeMB = 1024
		log.Printf("WARN: MAX_FILE_SIZE_MB not set or invalid, using default: %dMB", maxFileSizeMB)
	}

	return &Appconfig{
		Port:               port,
		Environment:        env,
		CouponFileURLs:     couponURLs,
		LocalCouponDirPath: localCouponDirPath,
		MaxFileSizeMB:      maxFileSizeMB,
	}
}
