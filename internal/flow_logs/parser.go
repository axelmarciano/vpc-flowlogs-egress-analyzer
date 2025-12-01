package flow_logs

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"vpc_flowlogs_egress_analyzer/internal/config"
)

var natEIPsCache []string

func getNatEIPs() []string {
	if natEIPsCache != nil {
		return natEIPsCache
	}
	env := config.GetEnv("NAT_EIPS_LIST")
	if env == "" {
		natEIPsCache = []string{}
		return natEIPsCache
	}
	parts := strings.Split(env, ",")
	natEIPsCache = make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			natEIPsCache = append(natEIPsCache, p)
		}
	}
	return natEIPsCache
}

func IsNatIP(ip string) bool {
	for _, n := range getNatEIPs() {
		if ip == n {
			return true
		}
	}
	return false
}

func ParseFlowLogLine(line string) (*VPCFlowLogRecord, error) {
	fields := strings.Fields(line)

	if len(fields) < 18 {
		return nil, fmt.Errorf("invalid flow log format: got %d fields, expected at least 18. Check README for required AWS Log format", len(fields))
	}

	v := VPCFlowLogRecord{}
	v.Version, _ = strconv.Atoi(fields[0])
	v.AccountId = fields[1]
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
	v.PktSrcAddr = fields[14]
	v.PktDstAddr = fields[15]
	v.PktSrcAwsService = fields[16]
	v.PktDstAwsService = fields[17]

	return &v, nil
}

func IsPrivateIP(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}
	if ip4 := ip.To4(); ip4 != nil {
		if ip4[0] == 10 {
			return true
		}
		if ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31 {
			return true
		}
		if ip4[0] == 192 && ip4[1] == 168 {
			return true
		}
		return false
	}
	if len(ip) == net.IPv6len && (ip[0]&0xfe) == 0xfc {
		return true
	}
	return false
}

func FlowDirection(r VPCFlowLogRecord) string {
	src := r.PktSrcAddr
	dst := r.PktDstAddr
	if dst == "" || src == "" {
		return "other"
	}

	srcIsPrivate := IsPrivateIP(src)
	dstIsPrivate := IsPrivateIP(dst)
	dstIsNatEIP := IsNatIP(dst)

	if srcIsPrivate && !dstIsPrivate && !dstIsNatEIP {
		return "egress"
	}

	if !srcIsPrivate && dstIsPrivate {
		return "ingress"
	}

	if srcIsPrivate && dstIsPrivate {
		return "internal"
	}

	return "other"
}
