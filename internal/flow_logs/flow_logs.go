package flow_logs

import (
	"bufio"
	"compress/gzip"
	"context"
	"fmt"
	"path"
	"strings"
	"vpc_flowlogs_egress_analyzer/internal/cache"

	services "vpc_flowlogs_egress_analyzer/internal/s3"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func RetrieveVPCFlowLogs() ([]VPCFlowLogRecord, error) {
	ctx := context.TODO()

	fmt.Println("Initializing S3 client...")
	s3Client, err := services.GetS3Client()
	if err != nil {
		return nil, fmt.Errorf("error getting S3 client: %w", err)
	}

	bucket, prefix, region, account, day, month, year, err := getFlowLogConfig()
	if err != nil {
		return nil, err
	}

	cacheKey := fmt.Sprintf("%s-%s-%s", year, month, day)

	fmt.Printf("➡ Selected date: %s-%s-%s\n", year, month, day)

	// CACHE CHECK
	if cache.Exists(cacheKey) {
		fmt.Println("📦 Loading VPC Flow Logs from cache:", cacheKey)
		return cache.Load[[]VPCFlowLogRecord](cacheKey)
	}

	fmt.Println("📦 No cache found, downloading logs from S3...")

	// BUILD S3 PREFIX PATH
	base := path.Join("AWSLogs", account, "vpcflowlogs", region, year, month, day)
	if prefix != "" {
		base = path.Join(prefix, base)
	}
	finalPrefix := base + "/"

	fmt.Printf("📁 Bucket: %s\n", bucket)
	fmt.Printf("📁 Prefix: %s\n", finalPrefix)

	input := &s3.ListObjectsV2Input{
		Bucket: &bucket,
		Prefix: &finalPrefix,
	}

	paginator := s3.NewListObjectsV2Paginator(s3Client, input)

	var logs []VPCFlowLogRecord
	totalFiles := 0
	totalLines := 0

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("list bucket error: %w", err)
		}

		fmt.Printf("📄 Found %d objects in this page\n", len(page.Contents))

		for _, obj := range page.Contents {
			if !strings.HasSuffix(*obj.Key, ".gz") {
				fmt.Printf("⏭ Skipping non-gzip file: %s\n", *obj.Key)
				continue
			}

			totalFiles++
			fmt.Printf("⬇ Downloading: %s\n", *obj.Key)

			out, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
				Bucket: &bucket,
				Key:    obj.Key,
			})
			if err != nil {
				return nil, fmt.Errorf("get object %s: %w", *obj.Key, err)
			}

			gzr, err := gzip.NewReader(out.Body)
			if err != nil {
				return nil, fmt.Errorf("gzip reader: %w", err)
			}

			scanner := bufio.NewScanner(gzr)
			linesInFile := 0

			for scanner.Scan() {
				line := scanner.Text()
				if strings.HasPrefix(line, "version") {
					continue
				}

				record, err := ParseFlowLogLine(line)
				if err != nil {
					fmt.Printf("⚠ Parse error on line: %v\n", err)
					continue
				}
				record.Direction = FlowDirection(*record)
				record.GB = float64(record.Bytes) / (1024 * 1024 * 1024)

				logs = append(logs, *record)
				linesInFile++
			}

			totalLines += linesInFile
			fmt.Printf("   ↳ Parsed %d lines\n", linesInFile)

			gzr.Close()
		}
	}

	fmt.Printf("✅ Total files processed: %d\n", totalFiles)
	fmt.Printf("✅ Total flow log lines parsed: %d\n", totalLines)

	// SAVE CACHE
	fmt.Printf("💾 Saving cache to .cache/%s.json.gz\n", cacheKey)
	_ = cache.Save(cacheKey, logs)

	fmt.Println("🎉 Logs retrieved successfully!")

	return logs, nil
}
