package flow_logs

type VPCFlowLogRecord struct {
	Version     int
	AccountID   string
	InterfaceID string
	SrcAddr     string
	DstAddr     string
	SrcPort     int
	DstPort     int
	Protocol    int
	Packets     int
	Bytes       int
	Start       int64
	End         int64
	Action      string
	LogStatus   string
	Direction   string  // egress, ingress, internal
	GB          float64 // bytes converted to gigabytes
}

type AnalysisSummary struct {
	Year  string `json:"year"`
	Month string `json:"month"`
	Day   string `json:"day"`

	ByIP map[string]*IPStats `json:"by_ip"`

	Total struct {
		Bytes   int     `json:"bytes"`
		GB      float64 `json:"gb"`
		CostUSD float64 `json:"cost_usd"`
	} `json:"total"`

	Region       string  `json:"region"`
	CostPerGBUSD float64 `json:"cost_per_gb_usd"`
}

type IPStats struct {
	Direction     string  `json:"direction"`
	Bytes         int     `json:"bytes"`
	GB            float64 `json:"gb"`
	CostUSD       float64 `json:"cost_usd"`
	ConnectionNum int     `json:"connection_count"`
}

type IPEntry struct {
	IP            string  `json:"ip"`
	Direction     string  `json:"direction"`
	Bytes         int     `json:"bytes"`
	GB            float64 `json:"gb"`
	CostUSD       float64 `json:"cost_usd"`
	ConnectionNum int     `json:"connection_count"`
}
