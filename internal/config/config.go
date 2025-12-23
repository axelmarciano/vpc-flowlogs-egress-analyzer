package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

func DefaultEnvValues() map[string]string {
	now := time.Now()
	year, month, day := now.Date()

	return map[string]string{
		"AWS_REGION":            "eu-west-3",
		"AWS_ACCESS_KEY_ID":     "",
		"AWS_SECRET_ACCESS_KEY": "",
		"S3_BUCKET_NAME":        "",
		"S3_PREFIX":             "",
		"AWS_ACCOUNT_ID":        "",
		"YEAR":                  fmt.Sprintf("%04d", year),
		"MONTH":                 fmt.Sprintf("%02d", int(month)),
		"DAY":                   fmt.Sprintf("%02d", day),
		"IP_INFO_API_KEY":       "",
		"NAT_EIPS_LIST":         "", // Comma-separated list of known NAT Gateway EIPs
	}
}

func LoadConfig() {
	err := godotenv.Load()
	if err != nil {
		log.Printf("No .env file found, continuing with runtime environment variables.")
	} else {
		log.Printf(".env file loaded")
	}
}

func GetEnv(key string) string {
	LoadConfig()
	value := os.Getenv(key)
	if value == "" {
		defaultValues := DefaultEnvValues()
		defaultValue := defaultValues[key]
		if defaultValue != "" {
			return defaultValue
		}
		return ""
	}
	return value
}
