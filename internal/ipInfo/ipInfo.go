package ipInfo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"vpc_flowlogs_egress_analyzer/internal/config"
)

func getApiKey() string {
	return config.GetEnv("IP_INFO_API_KEY")
}

type IpInfoResponse struct {
	IP             string `json:"ip"`
	ASN            string `json:"asn"`
	AS_NAME        string `json:"as_name"`
	AS_DOMAIN      string `json:"as_domain"`
	COUNTRY_CODE   string `json:"country_code"`
	COUNTRY        string `json:"country"`
	CONTINENT_CODE string `json:"continent_code"`
	CONTINENT      string `json:"continent"`
}

func GetIpInfo(ip string) (*IpInfoResponse, error) {
	apiKey := getApiKey()
	if apiKey == "" {
		return nil, nil
	}

	url := fmt.Sprintf("https://api.ipinfo.io/lite/%s?token=%s", ip, apiKey)

	client := http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("ipinfo http %d", resp.StatusCode)
	}

	var out IpInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}

	return &out, nil
}
