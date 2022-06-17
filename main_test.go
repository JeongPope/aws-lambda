package main

import (
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

func Test_convTimezone(t *testing.T) {
	type args struct {
		alertTime string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
		{"default", args{alertTime: "2022-06-01T12:08:19.123+0000"}, "2022-06-01 21:08:19.123 +0900 KST"},
		{"overDay", args{alertTime: "2022-06-01T15:08:19.456+0000"}, "2022-06-02 00:08:19.456 +0900 KST"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convTimezone(tt.args.alertTime); got != tt.want {
				t.Errorf("convTimezone() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetCause(t *testing.T) {
	type args struct {
		datas events.CloudWatchAlarmSNSPayload
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "Abnormal Bandwidth", args: args{
			datas: events.CloudWatchAlarmSNSPayload{
				AlarmName:        "[TEST] RDS CPU Utilization",
				AlarmDescription: "Test description",
				AWSAccountID:     "<MyAccount>",
				NewStateValue:    "ALARM",
				NewStateReason:   "reason",
				StateChangeTime:  "2022-06-01T15:08:19.456+0000",
				Region:           "AsiaPacific(Seoul)",
				AlarmARN:         "arn:aws:cloudwatch:ap-northeast-2:",
				OldStateValue:    "OK",
				Trigger: events.CloudWatchAlarmTrigger{
					Period:                           60,
					EvaluationPeriods:                10,
					ComparisonOperator:               "GreaterThanUpperThreshold",
					TreatMissingData:                 "- TreatMissingData:                    missing",
					EvaluateLowSampleCountPercentile: "",
					Metrics: []events.CloudWatchMetricDataQuery{
						{
							ID: "m1",
							MetricStat: events.CloudWatchMetricStat{
								Metric: events.CloudWatchMetric{
									Dimensions: []events.CloudWatchDimension{
										{
											Name:  "DBInstanceIdentifier",
											Value: "test-rds",
										},
									},
									MetricName: "CPUUtilization",
									Namespace:  "AWS/RDS",
								},
								Period: 60,
								Stat:   "Average",
							},
							ReturnData: true,
						},
						{
							ID:         "ad1",
							Label:      "[TEST] CPUUtilization",
							Expression: "ANOMALY_DETECTION_BAND(m1, 3)",
							ReturnData: true,
						},
					},
				},
			},
		}, want: "10 분 동안 10 회 | CPUUtilization 지표의 범위(약 3 배)를 벗어났습니다."},
		{name: "Threshold", args: args{
			datas: events.CloudWatchAlarmSNSPayload{
				AlarmName:        "[TEST] Network",
				AlarmDescription: "Test description",
				AWSAccountID:     "<MyAccount>",
				NewStateValue:    "ALARM",
				NewStateReason:   "reason",
				StateChangeTime:  "2022-06-01T15:08:19.456+0000",
				Region:           "AsiaPacific(Seoul)",
				AlarmARN:         "arn:aws:cloudwatch:ap-northeast-2:",
				OldStateValue:    "INSUFFICIENT_DATA",
				Trigger: events.CloudWatchAlarmTrigger{
					MetricName:    "NetworkOut",
					Namespace:     "AWS/EC2",
					StatisticType: "Statistic",
					Statistic:     "AVERAGE",
					Unit:          "Bytes",
					Dimensions: []events.CloudWatchDimension{
						{
							Value: "TestInstance",
							Name:  "InsatanceID",
						},
					},
					Period:                           60,
					EvaluationPeriods:                1,
					ComparisonOperator:               "GreaterThanThreshold",
					Threshold:                        0.0,
					TreatMissingData:                 "- TreatMissingData:                    missing",
					EvaluateLowSampleCountPercentile: "",
				},
			},
		}, want: "1 분 동안 1 회 | NetworkOut > 0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetCause(tt.args.datas); got != tt.want {
				t.Errorf("GetCause() = %v, want %v", got, tt.want)
			}
		})
	}
}

// func Test_abnormalBandMessage(t *testing.T) {
// 	type args struct {
// 		datas             events.CloudWatchAlarmSNSPayload
// 		evaluationPeriods int64
// 		minutes           float64
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 		want string
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if got := abnormalBandMessage(tt.args.datas, tt.args.evaluationPeriods, tt.args.minutes); got != tt.want {
// 				t.Errorf("abnormalBandMessage() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

// func Test_thresholdMessage(t *testing.T) {
// 	type args struct {
// 		datas             events.CloudWatchAlarmSNSPayload
// 		evaluationPeriods int64
// 		minutes           float64
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 		want string
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if got := thresholdMessage(tt.args.datas, tt.args.evaluationPeriods, tt.args.minutes); got != tt.want {
// 				t.Errorf("thresholdMessage() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }
