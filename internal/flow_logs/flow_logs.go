package flow_logs

import (
	"bufio"
	"compress/gzip"
	"context"
	"fmt"
	"path"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"vpc_flowlogs_egress_analyzer/internal/cache"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	services "vpc_flowlogs_egress_analyzer/internal/s3"
)

func RetrieveVPCFlowLogs() ([]VPCFlowLogRecord, error) {
	ctx := context.TODO()

	fmt.Println("Initializing S3 client…")
	s3Client, err := services.GetS3Client()
	if err != nil {
		return nil, err
	}

	bucket, prefix, region, account, day, month, year, err := getFlowLogConfig()
	if err != nil {
		return nil, err
	}

	cacheKey := fmt.Sprintf("%s-%s-%s", year, month, day)
	metaKey := cacheKey + "-meta"

	fmt.Printf("➡ Selected date: %s-%s-%s\n", year, month, day)

	if cache.Exists(metaKey) {
		fmt.Println("📦 Cache exists. Loading metadata…")
		meta, err := cache.Load[map[string]interface{}](metaKey)
		if err != nil {
			return nil, err
		}

		chunks := int(meta["chunks"].(float64))
		fmt.Printf("📦 Loading %d chunks in parallel…\n", chunks)

		numWorkers := runtime.NumCPU()
		if numWorkers < 2 {
			numWorkers = 2
		}
		fmt.Printf("⚙️ Using %d cache loader workers\n", numWorkers)

		type chunkResult struct {
			index int
			data  []VPCFlowLogRecord
			err   error
		}

		jobs := make(chan int, chunks)
		results := make(chan chunkResult, chunks)

		var wg sync.WaitGroup

		for w := 0; w < numWorkers; w++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()

				for idx := range jobs {
					fn := fmt.Sprintf("%s-part-%05d", cacheKey, idx)
					fmt.Printf("📥 Worker %d loading %s\n", workerID, fn)

					part, err := cache.Load[[]VPCFlowLogRecord](fn)
					results <- chunkResult{index: idx, data: part, err: err}
				}
			}(w)
		}

		for i := 1; i <= chunks; i++ {
			jobs <- i
		}
		close(jobs)

		go func() {
			wg.Wait()
			close(results)
		}()

		all := make([][]VPCFlowLogRecord, chunks)

		for r := range results {
			if r.err != nil {
				return nil, r.err
			}
			all[r.index-1] = r.data
		}

		var merged []VPCFlowLogRecord
		for _, part := range all {
			merged = append(merged, part...)
		}

		fmt.Printf("✅ Loaded %d flow records from cache\n", len(merged))
		return merged, nil
	}

	fmt.Println("📦 No cache found, downloading from S3…")

	base := path.Join("AWSLogs", account, "vpcflowlogs", region, year, month, day)
	if prefix != "" {
		base = path.Join(prefix, base)
	}
	finalPrefix := base + "/"

	fmt.Printf("📁 Bucket: %s\n", bucket)
	fmt.Printf("📁 Prefix: %s\n", finalPrefix)

	input := &s3.ListObjectsV2Input{Bucket: &bucket, Prefix: &finalPrefix}
	paginator := s3.NewListObjectsV2Paginator(s3Client, input)

	fmt.Println("🔍 Listing S3 objects…")

	var files []string
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, obj := range page.Contents {
			if strings.HasSuffix(*obj.Key, ".gz") {
				files = append(files, *obj.Key)
			}
		}
	}

	fmt.Printf("📄 Found %d gzip files\n", len(files))
	if len(files) == 0 {
		return nil, fmt.Errorf("no .gz files found")
	}

	numWorkers := runtime.NumCPU()
	numWriters := runtime.NumCPU() / 2
	if numWriters < 1 {
		numWriters = 1
	}

	fmt.Printf("⚙️ Using %d S3 workers\n", numWorkers)
	fmt.Printf("⚙️ Using %d writer goroutines\n", numWriters)

	fileCh := make(chan string, len(files))
	batchCh := make(chan []VPCFlowLogRecord, numWorkers*2)

	var wgWorkers sync.WaitGroup
	var wgWriters sync.WaitGroup

	var chunkIndex int64 = 0
	var total int64 = 0

	for i := 0; i < numWriters; i++ {
		wgWriters.Add(1)
		go func(writerID int) {
			defer wgWriters.Done()

			for batch := range batchCh {
				idx := atomic.AddInt64(&chunkIndex, 1)
				fn := fmt.Sprintf("%s-part-%05d", cacheKey, idx)

				fmt.Printf("💾 Writer %d saving %s (%d records)\n", writerID, fn, len(batch))
				if err := cache.Save(fn, batch); err != nil {
					fmt.Printf("❌ Writer %d error saving %s: %v\n", writerID, fn, err)
				}

				atomic.AddInt64(&total, int64(len(batch)))
			}
		}(i)
	}

	for w := 0; w < numWorkers; w++ {
		wgWorkers.Add(1)
		go func(workerID int) {
			defer wgWorkers.Done()

			local := make([]VPCFlowLogRecord, 0, 20000)

			for key := range fileCh {
				fmt.Printf("⬇ Worker %d downloading %s\n", workerID, key)

				out, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
					Bucket: &bucket,
					Key:    &key,
				})
				if err != nil {
					fmt.Printf("❌ Worker %d S3 error: %v\n", workerID, err)
					continue
				}

				gzr, err := gzip.NewReader(out.Body)
				if err != nil {
					fmt.Printf("❌ Worker %d gzip error: %v\n", workerID, err)
					continue
				}

				r := bufio.NewReaderSize(gzr, 256*1024)
				for {
					line, err := r.ReadBytes('\n')
					if err != nil {
						break
					}
					if strings.HasPrefix(string(line), "version") {
						continue
					}

					rec, err := ParseFlowLogLine(string(line))
					if err != nil {
						continue
					}

					rec.Direction = FlowDirection(*rec)
					rec.GB = float64(rec.Bytes) / (1024 * 1024 * 1024)

					local = append(local, *rec)

					if len(local) >= 20000 {
						b := make([]VPCFlowLogRecord, len(local))
						copy(b, local)
						batchCh <- b
						local = local[:0]
					}
				}

				gzr.Close()
				fmt.Printf("✔️ Worker %d finished %s\n", workerID, key)
			}

			if len(local) > 0 {
				b := make([]VPCFlowLogRecord, len(local))
				copy(b, local)
				batchCh <- b
			}
		}(w)
	}

	go func() {
		for _, key := range files {
			fileCh <- key
		}
		close(fileCh)
	}()

	go func() {
		wgWorkers.Wait()
		close(batchCh)
	}()

	wgWriters.Wait()

	fmt.Printf("📊 Total processed: %d records\n", total)

	meta := map[string]interface{}{
		"total":  total,
		"chunks": chunkIndex,
	}

	fmt.Println("💾 Saving metadata…")
	if err := cache.Save(metaKey, meta); err != nil {
		return nil, err
	}

	fmt.Println("📦 Reloading full logs from chunks…")

	var all []VPCFlowLogRecord
	for i := int64(1); i <= chunkIndex; i++ {
		fn := fmt.Sprintf("%s-part-%05d", cacheKey, i)
		fmt.Printf("📥 Loading %s\n", fn)
		part, err := cache.Load[[]VPCFlowLogRecord](fn)
		if err != nil {
			return nil, err
		}
		all = append(all, part...)
	}

	fmt.Printf("🎉 Completed. Total logs: %d\n", len(all))
	return all, nil
}
