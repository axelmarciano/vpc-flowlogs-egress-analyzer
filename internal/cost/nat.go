package cost

var NatDataProcessedCostPerGB = map[string]float64{
	"us-east-1": 0.045,
	"us-east-2": 0.045,
	"us-west-1": 0.055,
	"us-west-2": 0.045,

	"ca-central-1": 0.045,
	"ca-west-1":    0.055,

	"eu-west-1":    0.045,
	"eu-west-2":    0.052,
	"eu-west-3":    0.062,
	"eu-central-1": 0.052,
	"eu-central-2": 0.059,
	"eu-north-1":   0.048,
	"eu-south-1":   0.057,
	"eu-south-2":   0.057,

	"me-central-1": 0.066,
	"me-south-1":   0.066,

	"af-south-1": 0.066,

	"ap-south-1": 0.09,
	"ap-south-2": 0.09,

	"ap-northeast-1": 0.114,
	"ap-northeast-2": 0.106,
	"ap-northeast-3": 0.114,

	"ap-east-1": 0.114,

	"ap-southeast-1": 0.09,
	"ap-southeast-2": 0.095,
	"ap-southeast-3": 0.095,
	"ap-southeast-4": 0.106,
	"ap-southeast-5": 0.106,

	"sa-east-1": 0.12,
}
