package deepalert

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/pkg/errors"
)

// InspectHandler is a function type of callback of inspector.
type InspectHandler func(ctx context.Context, attr Attribute) (*TaskResult, error)

func publishSNS(topicArn string, data interface{}) error {
	// arn
	// aws
	// sns
	// ap-northeast-1
	// 789035092620
	// xxxxxx
	arr := strings.Split(topicArn, ":")
	if len(arr) != 6 {
		return fmt.Errorf("Invalid SNS ARN format: %s", topicArn)
	}
	region := arr[3]

	msg, err := json.Marshal(data)
	if err != nil {
		return errors.Wrap(err, "Fail to marshal report data")
	}

	ssn := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(region),
	}))
	snsService := sns.New(ssn)

	resp, err := snsService.Publish(&sns.PublishInput{
		Message:  aws.String(string(msg)),
		TopicArn: aws.String(topicArn),
	})

	if err != nil {
		return errors.Wrap(err, "Fail to publish report")
	}
	log.Printf("Published SNS %v to %s: %v", data, topicArn, resp)

	return nil
}

// StartInspector is a wrapper of Inspector.
func StartInspector(handler InspectHandler, author, submitTopic, attributeTopic string) {
	lambda.Start(func(ctx context.Context, event events.SNSEvent) error {
		for _, record := range event.Records {
			var task Task
			msg := record.SNS.Message
			if err := json.Unmarshal([]byte(msg), &task); err != nil {
				return errors.Wrapf(err, "Fail to unmarshal task: %s", msg)
			}

			result, err := handler(ctx, task.Attribute)
			if err != nil {
				return errors.Wrapf(err, "Fail to handle task: %v", task)
			}

			// Sending entities
			for _, entity := range result.Contents {
				content := ReportContent{
					ReportID:  task.ReportID,
					Attribute: task.Attribute,
					Author:    author,
					Type:      entity.Type(),
					Content:   entity,
				}

				if err := publishSNS(submitTopic, content); err != nil {
					return errors.Wrapf(err, "Fail to publish ReportContent to %s: %v", submitTopic, content)
				}
			}

			// Sending new attributes
			if len(result.NewAttributes) > 0 {
				attrReport := ReportAttribute{
					ReportID:   task.ReportID,
					Original:   task.Attribute,
					Attributes: result.NewAttributes,
					Author:     author,
				}

				if err := publishSNS(attributeTopic, attrReport); err != nil {
					return errors.Wrapf(err, "Fail to publish ReportAttribute to %s: %v", attributeTopic, attrReport)
				}
			}
		}

		return nil
	})
}