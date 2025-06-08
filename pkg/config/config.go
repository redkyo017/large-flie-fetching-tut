package config

import (
	"log"
	"os"
	"strconv"
)

type Appconfig struct {
	Port           string
	CouponFileURLs []string
	MaxFileSizeMB  int
}

func LoadConfig() *Appconfig {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	couponsURL := []string{
		"https://orderfoodonline-files.s3.ap-southeast-2.amazonaws.com/couponbase1.gz",
		"https://orderfoodonline-files.s3.ap-southeast-2.amazonaws.com/couponbase2.gz",
		"https://orderfoodonline-files.s3.ap-southeast-2.amazonaws.com/couponbase3.gz",
	}

	maxFileSizeMBStr := os.Getenv("MAX_FILE_SIZE_MB")
	maxFileSizeMB, err := strconv.Atoi(maxFileSizeMBStr)
	if err != nil || maxFileSizeMB <= 0 {
		maxFileSizeMB = 1024
		log.Printf("WARN: MAX_FILE_SIZE_MB not set or invalid, using default: %dMB", maxFileSizeMB)
	}

	return &Appconfig{
		Port:           port,
		CouponFileURLs: couponsURL,
		MaxFileSizeMB:  maxFileSizeMB,
	}
}
