package flow_logs

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"vpc_flowlogs_egress_analyzer/internal/cost"
	"vpc_flowlogs_egress_analyzer/internal/ipInfo"
)

func Analyze() {
	logs, err := RetrieveVPCFlowLogs()
	if err != nil {
		panic(fmt.Sprintf("CRITICAL: %v", err))
	}

	_, _, region, _, day, month, year, _ := getFlowLogConfig()

	summary := AnalysisSummary{
		Year:   year,
		Month:  month,
		Day:    day,
		ByIP:   make(map[string]*IPStats),
		Region: region,
	}

	costPerGB := cost.NatDataProcessedCostPerGB[region]
	summary.CostPerGBUSD = costPerGB

	var totalBytes int

	fmt.Println("🔍 Analyzing traffic patterns...")

	for _, r := range logs {
		if r.Direction != "egress" {
			continue
		}

		bytes := r.Bytes
		gb := float64(bytes) / (1024 * 1024 * 1024)
		costUSD := gb * costPerGB

		ip := r.PktDstAddr
		if ip == "" || ip == "-" {
			ip = r.DstAddr
		}

		if _, exists := summary.ByIP[ip]; !exists {
			summary.ByIP[ip] = &IPStats{Direction: "egress"}
		}

		stat := summary.ByIP[ip]
		stat.Bytes += bytes
		stat.GB += gb
		stat.CostUSD += costUSD
		stat.ConnectionNum++

		if r.PktDstAwsService != "-" && r.PktDstAwsService != "" {
			stat.AwsService = r.PktDstAwsService
		}

		totalBytes += bytes
	}

	summary.Total.Bytes = totalBytes
	summary.Total.GB = float64(totalBytes) / (1024 * 1024 * 1024)
	summary.Total.CostUSD = summary.Total.GB * costPerGB

	ips := make([]string, 0, len(summary.ByIP))
	for ip := range summary.ByIP {
		ips = append(ips, ip)
	}
	sort.Slice(ips, func(i, j int) bool {
		return summary.ByIP[ips[i]].GB > summary.ByIP[ips[j]].GB
	})

	topLimit := 50
	if len(ips) < topLimit {
		topLimit = len(ips)
	}

	fmt.Printf("🌍 Enriching top %d IPs with geo/ASN data...\n", topLimit)
	for i := 0; i < topLimit; i++ {
		ip := ips[i]
		info, err := ipInfo.GetIpInfo(ip)
		if err == nil {
			summary.ByIP[ip].IpInfo = info
		} else {
			log.Printf("⚠️ Warning: failed to get IpInfo for %s: %v", ip, err)
		}
	}

	saveAnalysisToFile(summary)
	printAnalysisSummary(summary)
}

func saveAnalysisToFile(summary AnalysisSummary) {
	entries := make([]IPEntry, 0, len(summary.ByIP))
	for ip, st := range summary.ByIP {
		entries = append(entries, IPEntry{
			IP:            ip,
			AwsService:    st.AwsService,
			Direction:     st.Direction,
			Bytes:         st.Bytes,
			GB:            st.GB,
			CostUSD:       st.CostUSD,
			ConnectionNum: st.ConnectionNum,
			IpInfo:        st.IpInfo,
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].GB > entries[j].GB
	})

	out := map[string]any{
		"year":            summary.Year,
		"month":           summary.Month,
		"day":             summary.Day,
		"region":          summary.Region,
		"cost_per_gb_usd": summary.CostPerGBUSD,
		"total": map[string]any{
			"bytes":    summary.Total.Bytes,
			"gb":       summary.Total.GB,
			"cost_usd": summary.Total.CostUSD,
		},
		"egress_by_ip": entries,
	}

	j, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		fmt.Println("❌ Error marshaling JSON:", err)
		return
	}

	err = os.WriteFile("result.json", j, 0644)
	if err != nil {
		fmt.Println("❌ Error writing result.json:", err)
		return
	}

	fmt.Println("💾 Saved detailed analysis to result.json")
}

func printAnalysisSummary(s AnalysisSummary) {
	fmt.Println("\n=================================================================")
	fmt.Printf("📊 VPC Egress Cost Analysis | %s-%s-%s | %s\n", s.Year, s.Month, s.Day, s.Region)
	fmt.Println("=================================================================")

	totalIPs := len(s.ByIP)

	fmt.Printf("💰 Total Estimated NAT Cost:   $%.2f\n", s.Total.CostUSD)
	fmt.Printf("📡 Total Data Processed:       %.2f GB\n", s.Total.GB)
	fmt.Printf("🎯 Unique Destination IPs:     %d\n", totalIPs)

	fmt.Println("-----------------------------------------------------------------")
	fmt.Println("💡 Optimization Hint: Look for 'S3' or 'DYNAMODB' in result.json")
	fmt.Println("   Use Gateway Endpoints (free) instead of NAT (paid) for these.")
	fmt.Println("=================================================================")
}
