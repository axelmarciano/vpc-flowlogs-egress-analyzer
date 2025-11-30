package main

import (
	"vpc_flowlogs_egress_analyzer/internal/config"
	"vpc_flowlogs_egress_analyzer/internal/flow_logs"
)

func main() {
	config.LoadConfig()
	flow_logs.Analyze()
}
