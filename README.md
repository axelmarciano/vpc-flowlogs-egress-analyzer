> **Stop burning money on AWS NAT Gateways.**
> A high-performance, parallel Go tool to analyze VPC Flow Logs, visualize Egress traffic, and uncover hidden costs.

![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)
![License](https://img.shields.io/badge/License-MIT-green.svg)
![AWS](https://img.shields.io/badge/AWS-VPC%20Flow%20Logs-orange?style=flat&logo=amazon-aws)

**VPC Flow Logs Egress Analyzer** acts as a detective for your AWS network bill. It downloads gigabytes of logs, parses them in seconds, and tells you exactly **who** is talking to the internet and **how much** it costs.

It specifically hunts for **"Wasteful Traffic"**: instances talking to AWS services (S3, DynamoDB) via a NAT Gateway instead of free VPC Endpoints.

---

## 🚨 Critical Prerequisite: Log Configuration

**This tool requires packet-level visibility to function.** The default AWS Flow Log format is insufficient because it masks the true source of traffic behind the NAT IP.

You **must** configure your VPC Flow Logs with a **Custom Format**.

### How to set it up:
1.  Go to the **AWS VPC Console** > **Your VPCs**.
2.  Select your VPC > **Flow Logs** tab > **Create flow log**.
3.  Under **Log record format**, choose **Custom format**.
4.  **Copy and paste this EXACT string** into the format box (order matters):

```text
\${version} \${account-id} \${interface-id} \${srcaddr} \${dstaddr} \${srcport} \${dstport} \${protocol} \${packets} \${bytes} \${start} \${end} \${action} \${log-status} \${pkt-srcaddr} \${pkt-dstaddr} \${pkt-src-aws-service} \${pkt-dst-aws-service}
```

> **Why?**
> * `${pkt-srcaddr}` / `${pkt-dstaddr}`: Reveals the *original* IP before it was NAT-ed.
> * `${pkt-dst-aws-service}`: Tells us if you are paying NAT fees to talk to S3, DynamoDB, or Kinesis.

---

## ✨ Key Features

### 🚀 High-Performance Architecture
Processing 50GB of logs? No problem.
* **Parallel S3 Downloader**: Fetches logs in chunks.
* **Streaming Gzip Parser**: Decompresses on the fly.
* **One-Shot Analysis**: No database required.

### 🧠 Smart Traffic Classification
It doesn't just look at IPs; it understands network flow directions:
* **Egress**: Traffic leaving your private subnet $\\rightarrow$ Internet (Expensive 💸).
* **Ingress**: Internet $\\rightarrow$ Public Subnet.
* **Internal**: Private $\\rightarrow$ Private.

### 💰 Cost Optimization Engine
* **AWS Service Detection**: Identifies traffic to AWS services (S3, DynamoDB) passing through NAT.
* **Cost Calculator**: Estimates `Data Processed` fees based on the region's pricing.
* **Enrichment**: Top 50 IPs are enriched with **ASN, ISP, and Country** data via IpInfo.

### 📂 Efficient Caching
Includes a local file cache (`.cache/`). Re-running the tool on the same day is instant.

---

## 🛠️ Installation & Usage

### 🐳 Using Docker (Recommended)

No Go installation required. Just map your credentials and run.

```bash
# 1. Build the image
make build docker

# 2. Run the analyzer
make run docker
```

### ⚙️ Environment Variables

Create a `.env` file or pass these to Docker:

| Variable | Required | Description |
| :--- | :---: | :--- |
| `S3_BUCKET_NAME` | ✅ | The bucket where Flow Logs are stored. |
| `AWS_ACCOUNT_ID` | ✅ | Used to locate the logs in the S3 path. |
| `AWS_REGION` | ❌ | Region to analyze (default: current region). |
| `S3_PREFIX` | ❌ | Custom prefix if logs aren't in root. |
| `YEAR` / `MONTH` / `DAY` | ❌ | Date to analyze (default: today). |
| `NAT_EIPS_LIST` | ❌ | Comma-separated list of your NAT Gateway Elastic IPs (helps filter noise). |
| `IP_INFO_API_KEY` | ❌ | Your token from ipinfo.io (for better geo-data). |

---

## 📊 Understanding the Output

The tool generates a detailed `result.json` and prints a summary.

### Example Console Output
```text
=================================================================
📊 VPC Egress Cost Analysis | 2025-12-01 | eu-west-3
=================================================================
💰 Total Estimated NAT Cost:   $34.20
📡 Total Data Processed:       760.44 GB
🎯 Unique Destination IPs:     4,210
-----------------------------------------------------------------
💡 Optimization Hint: Look for 'S3' or 'DYNAMODB' in result.json
Use Gateway Endpoints (free) instead of NAT (paid) for these.
=================================================================
```

### Example `result.json` (The Action Plan)

This JSON tells you exactly where to optimize.

```json
{
"total": {
"gb": 55.4,
"cost_usd": 2.49
},
"egress_by_ip": [
{
"ip": "52.218.x.x",
"aws_service": "S3",          
"direction": "egress",
"bytes": 53687091200,
"gb": 50.0,
"cost_usd": 2.25,
"ipinfo": { "as_name": "Amazon.com, Inc." }
},
{
"ip": "1.1.1.1",
"aws_service": "",
"direction": "egress",
"gb": 0.5,
"cost_usd": 0.02
}
]
}
```

## 📉 How to Interpret & Fix

1.  **Found `aws_service: "S3"` or `"DYNAMODB"`?**
    * **Problem:** Your app is accessing S3 buckets via the NAT Gateway. You are paying ~$0.045/GB.
    * **Fix:** Create a **VPC Gateway Endpoint** for S3/DynamoDB in your route table.
    * **Savings:** 100% of that cost becomes **$0**.

2.  **Found `aws_service: "AMAZON"` (EC2, etc)?**
    * **Problem:** Talking to AWS APIs or other regions via Public Internet.
    * **Fix:** Consider using **VPC Interface Endpoints (PrivateLink)**.

3.  **High Traffic to unknown Public IPs?**
    * **Action:** Check the `ipinfo` field. Is it a 3rd party API? A monitoring tool? Is the volume expected?

---

## 🔐 Authentication

The tool supports the standard AWS SDK credential chain:
1.  Environment Variables (`AWS_ACCESS_KEY_ID`...)
2.  `~/.aws/credentials` (mounted in Docker)
3.  IAM Roles (if running on EC2/EKS)
4.  AWS SSO profiles

---

## 📜 License

MIT License. Built with ❤️ to save you 💰.
