package flow_logs

import (
	"fmt"
	"strings"
	"vpc_flowlogs_egress_analyzer/internal/config"
)

// Read config automatically
func getFlowLogConfig() (bucket, prefix, region, account, day, month, year string, err error) {
	bucket = config.GetEnv("S3_BUCKET_NAME")
	if bucket == "" {
		return "", "", "", "", "", "", "", fmt.Errorf("missing env: S3_BUCKET")
	}

	region = config.GetEnv("AWS_REGION")
	if region == "" {
		return "", "", "", "", "", "", "", fmt.Errorf("missing env: AWS_REGION")
	}

	account = config.GetEnv("AWS_ACCOUNT_ID")
	if account == "" {
		return "", "", "", "", "", "", "", fmt.Errorf("missing env: AWS_ACCOUNT_ID")
	}
	day = config.GetEnv("DAY")
	month = config.GetEnv("MONTH")
	year = config.GetEnv("YEAR")

	prefix = config.GetEnv("S3_PREFIX") // optional
	prefix = strings.Trim(prefix, "/")

	return bucket, prefix, region, account, day, month, year, nil
}
