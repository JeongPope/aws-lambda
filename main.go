package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type statusMessage struct {
	Color   string
	Message string
}

var (
	slack_webhook = os.Getenv("SLACK_INCOMING_WEBHOOK")

	statusTypes = map[string]statusMessage{
		"ALARM":             {Color: "danger", Message: "위험"},
		"INSUFFICIENT_DATA": {Color: "warning", Message: "데이터 부족"},
		"OK":                {Color: "good", Message: "정상"},
	}

	operator = map[string]string{
		"GreaterThanOrEqualToThreshold": ">=",
		"GreaterThanThreshold":          ">",
		"LowerThanOrEqualToThreshold":   "<=",
		"LessThanThreshold":             "<",
	}
)

func main() {
	lambda.Start(handler)
}

func handler(context context.Context, snsEvent events.SNSEvent) (string, error) {
	snsMessage := snsEvent.Records[0].SNS.Message

	postJsonMessage := BuildSlackBlockMessage(snsMessage)
	err := PostSlack(postJsonMessage)

	if err != nil {
		return "Failed", err
	}

	return "Done", nil
}

func BuildSlackBlockMessage(snsMessage string) []byte {
	var datas events.CloudWatchAlarmSNSPayload

	err := json.Unmarshal([]byte(snsMessage), &datas)
	if err != nil {
		log.Fatalln(err.Error())
	}

	currState := statusTypes[datas.NewStateValue]
	prevState := statusTypes[datas.OldStateValue]
	kstTime := convTimezone(datas.StateChangeTime)
	description := datas.AlarmDescription
	cause := GetCause(datas)

	type ItemField struct {
		Title string `json:"title"`
		Value string `json:"value"`
		Short bool   `json:"short,omitempty"`
	}

	type AttachmentItems struct {
		Title  string      `json:"title"`
		Color  string      `json:"color"`
		Fields []ItemField `json:"fields"`
	}

	type Attachments struct {
		Items []AttachmentItems `json:"attachments"`
	}

	message := Attachments{}
	message.Items = append(message.Items, AttachmentItems{})
	message.Items[0].Title = datas.AlarmName
	message.Items[0].Color = currState.Color
	message.Items[0].Fields = append(message.Items[0].Fields, ItemField{Title: "When", Value: kstTime})
	message.Items[0].Fields = append(message.Items[0].Fields, ItemField{Title: "Desc", Value: description})
	message.Items[0].Fields = append(message.Items[0].Fields, ItemField{Title: "Cause", Value: cause})
	message.Items[0].Fields = append(message.Items[0].Fields, ItemField{Title: "Prev State", Value: prevState.Message, Short: true})
	message.Items[0].Fields = append(message.Items[0].Fields, ItemField{Title: "Curr State", Value: currState.Message, Short: true})
	message.Items[0].Fields = append(message.Items[0].Fields, ItemField{Title: "Link", Value: createCloudwatchLink(datas.AlarmARN, datas.AlarmName)})

	jsonMessage, _ := json.Marshal(message)

	return jsonMessage
}

func convTimezone(alertTime string) string {
	if alertTime == "" {
		return ""
	}

	replaceStr := strings.Replace(alertTime, "+0000", "Z", -1)
	parseTime, err := time.Parse(time.RFC3339Nano, replaceStr)
	if err != nil {
		panic(err)
	}

	loc, err := time.LoadLocation("Asia/Seoul")
	if err != nil {
		panic(err)
	}

	kstTime := parseTime.In(loc)

	return kstTime.String()
}

func GetCause(datas events.CloudWatchAlarmSNSPayload) string {
	trigger := datas.Trigger
	evaluationPeriods := trigger.EvaluationPeriods
	minutes := math.Floor(float64(trigger.Period) / 60)

	if len(trigger.Metrics) > 0 {
		return abnormalBandMessage(datas, evaluationPeriods, minutes)
	}

	return thresholdMessage(datas, evaluationPeriods, minutes)
}

func abnormalBandMessage(datas events.CloudWatchAlarmSNSPayload,
	evaluationPeriods int64, minutes float64) string {
	metrics := datas.Trigger.Metrics

	var metricName, expression string
	for _, metric := range metrics {
		if metric.ID == "m1" {
			metricName = metric.MetricStat.Metric.MetricName
		}

		if metric.ID == "ad1" {
			expression = metric.Expression
		}
	}

	var width string
	width = strings.Split(expression, ",")[1]
	width = strings.Replace(width, ")", "", -1)
	width = strings.Replace(width, " ", "", -1)

	return fmt.Sprintf("%d 분 동안 %d 회 | %s 지표의 범위(약 %s 배)를 벗어났습니다.",
		evaluationPeriods*int64(minutes), evaluationPeriods,
		metricName, width)
}

func thresholdMessage(datas events.CloudWatchAlarmSNSPayload,
	evaluationPeriods int64, minutes float64) string {
	trigger := datas.Trigger
	threshold := trigger.Threshold
	metricName := trigger.MetricName
	oper := operator[trigger.ComparisonOperator]
	fmt.Println(oper)

	return fmt.Sprintf("%d 분 동안 %d 회 | %s %s %d",
		evaluationPeriods*int64(minutes), evaluationPeriods,
		metricName, oper, int64(threshold))
}

func createCloudwatchLink(alarmArn, alarmName string) string {
	region := strings.Replace(alarmArn, "arn:aws:cloudwatch:", "", -1)
	region = strings.Split(region, ":")[0]

	escapeCode := url.QueryEscape(alarmName)

	return "https://ap-northeast-2.console.aws.amazon.com/cloudwatch/home" +
		"?region=" + region +
		"#alarmsV2:alarm/" + escapeCode
}

func PostSlack(jsonBody []byte) error {
	buffer := bytes.NewBuffer(jsonBody)

	resp, err := http.Post(slack_webhook, "application/json", buffer)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(string(respBody))
	}

	return nil
}
