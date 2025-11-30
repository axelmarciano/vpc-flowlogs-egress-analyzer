📊 VPC Flow Logs Egress Analyzer

VPC Flow Logs Egress Analyzer is a lightweight, one-shot Go tool designed to:
•	Fetch AWS VPC Flow Logs directly from S3
•	Parse and classify traffic as egress, ingress, or internal
•	Aggregate data per destination IP (only for egress)
•	Calculate the AWS NAT Gateway Data Processed cost based on region
•	Export a clean, sorted summary as result.json
•	Display a human-readable summary in the console
•	Use a local cache to avoid re-downloading and re-parsing logs

This tool is ideal for identifying:
•	Unexpected outbound traffic
•	NAT Gateway cost leaks
•	External services your infrastructure is communicating with
•	Daily egress volume per IP


⸻

✨ Features

✔ Retrieve VPC Flow Logs directly from S3

Uses your AWS Account ID, region, and bucket structure:

```
AWSLogs/<ACCOUNT_ID>/vpcflowlogs/<region>/<year>/<month>/<day>/
```

✔ Full VPC Flow Log parser (AWS format V2)

✔ Egress/Internal/Ingress detection

Automatic classification based on IP ranges.

✔ Per-IP aggregation

For each external IP:
•	total bytes
•	GB processed
•	NAT cost (region-aware)
•	connection count

✔ NAT Gateway cost estimation

Uses official AWS NAT Gateway Data Processed pricing, region-specific.

✔ Smart caching

Downloaded logs are written to .cache/YYYY-MM-DD.json.gz
→ next runs are instant.

✔ Clean final output
•	result.json is sorted by GB DESC
•	Console summary is compact and readable

✔ AWS credentials auto-loaded

No need to hardcode credentials, see below.

⸻

⚙️ Environment Variables

The following variables configure the tool.
They are automatically loaded with defaults using:

```
func DefaultEnvValues() map[string]string {
	now := time.Now()
	year, month, day := now.Date()

	return map[string]string{
		"AWS_REGION":            "eu-west-3",
		"AWS_ACCESS_KEY_ID":     "",
		"AWS_SECRET_ACCESS_KEY": "",
		"S3_BUCKET_NAME":        "",
		"S3_PREFIX":             "",
		"AWS_ACCOUNT_ID":        "",
		"YEAR":                  fmt.Sprintf("%04d", year),
		"MONTH":                 fmt.Sprintf("%02d", int(month)),
		"DAY":                   fmt.Sprintf("%02d", day),
	}
}
```

| Variable               | Description                                      | Default Value      |
|------------------------|--------------------------------------------------|--------------------|
| AWS_REGION             | AWS region where the VPC Flow Logs are stored    | eu-west-3          |
| AWS_ACCESS_KEY_ID      | Your AWS Access Key ID                           | (empty)             |
| AWS_SECRET_ACCESS_KEY  | Your AWS Secret Access Key                       | (empty)             |
| S3_BUCKET_NAME         | Name of the S3 bucket containing the Flow Logs   | (empty)             |
| S3_PREFIX              | Prefix path in the S3 bucket                        | (empty)             |
| AWS_ACCOUNT_ID         | Your AWS Account ID                              | (empty)             |
| YEAR                   | Year of the logs to fetch (YYYY)                 | Current year       |
| MONTH                  | Month of the logs to fetch (MM)                  | Current month      |
| DAY                    | Day of the logs to fetch (DD)                    | Current day        |

⸻

🔐 AWS Authentication

You do not need to provide AWS_ACCESS_KEY_ID or AWS_SECRET_ACCESS_KEY if:

✔ You run using Docker and mount your local AWS credentials:

```
-v ~/.aws:/root/.aws:ro
```


✔ You are authenticated via:
•	AWS CLI v2 SSO
•	MFA sessions
•	Named profiles
•	~/.aws/credentials or ~/.aws/config

The tool will automatically pick up:
•	your default profile
•	or whatever profile is configured in your shell

No credentials are ever stored by the tool.

⸻

🗄 Caching Behavior

The tool creates a .cache/ folder (gitignored):

Cache contains:
•	fully parsed VPC Flow Logs
•	enriched records (direction, GB, cost, etc.)

If a cache file already exists for the selected date:
→ S3 is not queried
→ Logs are loaded instantly from disk

You can safely delete .cache/ at any time.

⸻

📦 Output: result.json

A clean JSON file is generated:
•	sorted by GB descending
•	including full cost breakdown
•	per-IP details
•	total cost summary

Example result.json:

```json
{
  "year": "2025",
  "month": "11",
  "day": "30",
  "region": "eu-west-3",
  "cost_per_gb_usd": 0.062,
  "total": {
    "bytes": 5772308394,
    "gb": 5.382,
    "cost_usd": 0.3337
  },
  "egress_by_ip": [
    {
      "ip": "3.5.204.12",
      "direction": "egress",
      "bytes": 1090519040,
      "gb": 1.015625,
      "cost_usd": 0.063,
      "connection_count": 120
    }
  ]
}
```

⸻

🐳 Running with Docker

The Makefile handles everything.

Build the image:

```
make build docker
```

Run analysis:
```
make run docker
```


