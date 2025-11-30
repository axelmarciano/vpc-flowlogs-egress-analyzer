package flow_logs

import (
	"fmt"
	"strconv"
	"strings"
)

func ParseFlowLogLine(line string) (*VPCFlowLogRecord, error) {
	fields := strings.Fields(line)
	if len(fields) < 14 { // some regions add fields, but minimum is 14
		return nil, fmt.Errorf("invalid flow log line (too short): %s", line)
	}

	v := VPCFlowLogRecord{}

	var err error

	v.Version, _ = strconv.Atoi(fields[0])
	v.AccountID = fields[1]
	v.InterfaceID = fields[2]
	v.SrcAddr = fields[3]
	v.DstAddr = fields[4]
	v.SrcPort, _ = strconv.Atoi(fields[5])
	v.DstPort, _ = strconv.Atoi(fields[6])
	v.Protocol, _ = strconv.Atoi(fields[7])
	v.Packets, _ = strconv.Atoi(fields[8])
	v.Bytes, _ = strconv.Atoi(fields[9])
	v.Start, _ = strconv.ParseInt(fields[10], 10, 64)
	v.End, _ = strconv.ParseInt(fields[11], 10, 64)
	v.Action = fields[12]
	v.LogStatus = fields[13]

	return &v, err
}

func IsPrivateIP(ip string) bool {
	privateIPBlocks := []string{
		"10.",
		"172.16.", "172.17.", "172.18.", "172.19.",
		"172.20.", "172.21.", "172.22.", "172.23.",
		"172.24.", "172.25.", "172.26.", "172.27.",
		"172.28.", "172.29.", "172.30.", "172.31.",
		"192.168.",
		"fd", // IPv6 ULA
	}

	for _, block := range privateIPBlocks {
		if strings.HasPrefix(ip, block) {
			return true
		}
	}
	return false
}

func FlowDirection(r VPCFlowLogRecord) string {
	privateSrc := IsPrivateIP(r.SrcAddr)
	privateDst := IsPrivateIP(r.DstAddr)

	switch {
	case privateSrc && !privateDst:
		return "egress"
	case !privateSrc && privateDst:
		return "ingress"
	case privateSrc:
		return "internal"
	default:
		return "external"
	}
}
