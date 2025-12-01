package flow_logs

import "vpc_flowlogs_egress_analyzer/internal/ipInfo"

// VPCFlowLogRecord correspond exactement au format Custom attendu
type VPCFlowLogRecord struct {
	Version          int
	AccountId        string
	InterfaceID      string
	SrcAddr          string
	DstAddr          string
	SrcPort          int
	DstPort          int
	Protocol         int
	Packets          int
	Bytes            int
	Start            int64
	End              int64
	Action           string
	LogStatus        string
	PktSrcAddr       string
	PktDstAddr       string
	PktSrcAwsService string
	PktDstAwsService string
	Direction        string
}

type AnalysisSummary struct {
	Year         string  `json:"year"`
	Month        string  `json:"month"`
	Day          string  `json:"day"`
	Region       string  `json:"region"`
	CostPerGBUSD float64 `json:"cost_per_gb_usd"`

	Total struct {
		Bytes   int     `json:"bytes"`
		GB      float64 `json:"gb"`
		CostUSD float64 `json:"cost_usd"`
	} `json:"total"`

	ByIP map[string]*IPStats `json:"-"`
}

type IPStats struct {
	Direction     string
	Bytes         int
	GB            float64
	CostUSD       float64
	ConnectionNum int
	AwsService    string
	IpInfo        *ipInfo.IpInfoResponse
}

// IPEntry est la structure finale pour le JSON
type IPEntry struct {
	IP            string  `json:"ip"`
	AwsService    string  `json:"aws_service,omitempty"`
	Direction     string  `json:"direction"`
	Bytes         int     `json:"bytes"`
	GB            float64 `json:"gb"`
	CostUSD       float64 `json:"cost_usd"`
	ConnectionNum int     `json:"connection_num"`
	IpInfo        any     `json:"ipinfo"`
}
