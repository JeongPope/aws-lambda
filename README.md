# aws-lambda
* Cloudwatch에서 이상 지표가 탐지되는것을 Slack 메세지로 보냅니다
* Cloudwatch Alarm -> AWS SNS Event -> Lambda -> Slack incoming webhook

### zip
```bash
$ GOARCH=amd64 GOOS=linux go build main.go
$ zip {{name}}.zip main
```

### AWS Lambda : Runtime
* Lambda Runtime 설정에서 핸들러는 `handler` 함수가 아닌 `main` 입니다.


### Reference
* [AWS Official Docs](https://docs.aws.amazon.com/ko_kr/lambda/latest/dg/golang-handler.html)
* [AWS Official Github : Events](https://github.com/aws/aws-lambda-go/tree/main/events)
* [Tistory : jojoldu](https://jojoldu.tistory.com/586)