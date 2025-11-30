package flow_logs

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"vpc_flowlogs_egress_analyzer/internal/cost"
)

func Analyze() {
	logs, err := RetrieveVPCFlowLogs()
	if err != nil {
		panic(err)
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

	for _, r := range logs {
		if r.Direction != "egress" {
			continue
		}

		bytes := r.Bytes
		gb := float64(bytes) / (1024 * 1024 * 1024)
		costUSD := gb * costPerGB

		ip := r.DstAddr

		if _, exists := summary.ByIP[ip]; !exists {
			summary.ByIP[ip] = &IPStats{
				Direction: "egress",
			}
		}

		stat := summary.ByIP[ip]
		stat.Bytes += bytes
		stat.GB += gb
		stat.CostUSD += costUSD
		stat.ConnectionNum++

		totalBytes += bytes
	}

	summary.Total.Bytes = totalBytes
	summary.Total.GB = float64(totalBytes) / (1024 * 1024 * 1024)
	summary.Total.CostUSD = summary.Total.GB * costPerGB

	saveAnalysisToFile(summary)

	printAnalysisSummary(summary)
}

func saveAnalysisToFile(summary AnalysisSummary) {
	entries := make([]IPEntry, 0, len(summary.ByIP))
	for ip, st := range summary.ByIP {
		entries = append(entries, IPEntry{
			IP:            ip,
			Direction:     st.Direction,
			Bytes:         st.Bytes,
			GB:            st.GB,
			CostUSD:       st.CostUSD,
			ConnectionNum: st.ConnectionNum,
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

	fmt.Println("💾 Saved analysis to result.json (sorted by GB desc)")
}

func printAnalysisSummary(s AnalysisSummary) {
	fmt.Println("=================================================================")
	fmt.Printf("VPC Flow Logs Summary for %s-%s-%s (Region: %s)\n",
		s.Year, s.Month, s.Day, s.Region)
	fmt.Println("=================================================================")

	totalIPs := len(s.ByIP)

	fmt.Printf(
		"➡️  Egressed to %d unique IPs | Total: %.3f GB | Estimated NAT Cost: $%.4f\n",
		totalIPs,
		s.Total.GB,
		s.Total.CostUSD,
	)

	fmt.Printf("   NAT Pricing: %.3f $/GB\n", s.CostPerGBUSD)

	fmt.Println("=================================================================")
}
