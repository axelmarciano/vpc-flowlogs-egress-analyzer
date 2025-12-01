# 📊 VPC Flow Logs Egress Analyzer

VPC Flow Logs Egress Analyzer is a fast, parallel, one-shot Go tool designed to analyze AWS VPC Flow Logs and extract meaningful insights about outbound traffic and NAT Gateway costs.

It automatically:

- Fetches AWS VPC Flow Logs from S3
- Parses and classifies each flow as **egress / ingress / internal**
- Aggregates egress traffic **per destination IP**
- Calculates AWS **NAT Gateway Data Processed** cost
- Saves a clean `result.json`
- Displays a readable summary
- Uses a smart local cache to avoid re-downloading logs
- Optionally enriches **the top 50 IPs** with ASN/country/continent via **IpInfo**

---

## ✨ Features

### ✔ High-performance S3 log processing
Downloads and parses all VPC Flow Logs from:

```
AWSLogs/<ACCOUNT_ID>/vpcflowlogs/<region>/<year>/<month>/<day>/
```

Processing uses multiple worker pools:
- S3 download workers
- Gzip parsing workers
- Disk writer workers

### ✔ Full AWS VPC Flow Logs v2 parser

### ✔ Traffic classification

- egress
- ingress
- internal

### ✔ Per-IP aggregation

Each destination IP receives aggregated metrics:
- Total bytes
- Total GB
- NAT cost
- Connection count
- **IpInfo enrichment (only top 50 IPs)**

### ✔ NAT Gateway cost estimation

### ✔ Smart caching

Cached files are stored under `.cache/`.  
If present, logs load instantly.

### ✔ Output

- `result.json` sorted by GB descending
- Console summary

---

## ⚙️ Environment Variables

```
AWS_REGION
AWS_ACCESS_KEY_ID
AWS_SECRET_ACCESS_KEY
S3_BUCKET_NAME
S3_PREFIX
AWS_ACCOUNT_ID
YEAR
MONTH
DAY
IP_INFO_API_KEY (not required, only for IpInfo enrichment)
```

Defaults match the current date and region.

---

## 🔐 AWS Authentication

Supports:
- AWS CLI v2
- SSO
- ~/.aws/credentials
- Role-based auth
- Docker-mounted credentials

---

## 📦 Output Example

```json
{
  "year": "2025",
  "month": "12",
  "day": "01",
  "region": "eu-west-3",
  "cost_per_gb_usd": 0.062,
  "total": {
    "bytes": 5772308394,
    "gb": 5.382,
    "cost_usd": 0.3337
  },
  "egress_by_ip": [
    {
      "ip": "XX.XX.XX.XX",
      "direction": "egress",
      "bytes": 1090519040,
      "gb": 1.015625,
      "cost_usd": 0.063,
      "connection_count": 120,
      "ipinfo": {
        "ip": "XX.XX.XX.XX",
        "asn": "AS16509",
        "as_name": "Amazon.com, Inc.",
        "as_domain": "amazon.com",
        "country_code": "FR",
        "country": "France",
        "continent_code": "EU",
        "continent": "Europe"
      }
    }
  ]
}
```

---

## 🐳 Docker Usage

```
make build
make run
```
