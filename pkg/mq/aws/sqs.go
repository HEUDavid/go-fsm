package aws

import (
	"context"
	"fmt"
	"github.com/HEUDavid/go-fsm/pkg/mq"
	"github.com/HEUDavid/go-fsm/pkg/util"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"log"
	"time"
)

type Factory struct {
	Section string
	buffer  chan mq.Message
	queue   string
	sqs     *sqs.SQS
}

func (f *Factory) GetMQSection() string {
	return f.Section
}

func (f *Factory) InitMQ(config util.Config) error {
	f.queue = config["queue"].(string)

	accessKey := config["accessKey"].(string)
	secretKey := config["secretKey"].(string)
	region := config["region"].(string)

	sess, err := session.NewSession(&aws.Config{
		Region:      &region,
		Credentials: credentials.NewStaticCredentials(accessKey, secretKey, ""),
	})
	if err != nil {
		return err
	}

	f.sqs = sqs.New(sess)

	f.buffer = make(chan mq.Message)

	return nil
}

func (f *Factory) PublishMessage(c context.Context, msg string) error {
	if _, err := f.sqs.SendMessage(&sqs.SendMessageInput{
		MessageBody: &msg,
		QueueUrl:    &f.queue,
	}); err != nil {
		return err
	}
	return nil
}

func (f *Factory) FetchMessage(c context.Context) mq.Message {
	msg, ok := <-f.buffer
	if !ok {
		return mq.Message{C: c}
	}
	return msg
}

func (f *Factory) Start() {
	go func() {
		for {
			result, err := f.sqs.ReceiveMessage(&sqs.ReceiveMessageInput{
				QueueUrl:            &f.queue,
				MaxNumberOfMessages: aws.Int64(1),
				WaitTimeSeconds:     aws.Int64(20),
			})
			if err != nil {
				log.Printf("[FSM] Error receiving message: %v", err)
				continue
			}

			for _, message := range result.Messages {
				f.buffer <- mq.Message{
					C:    context.Background(),
					Body: *message.Body,

					Ack: func() error {
						if _, e := f.sqs.DeleteMessage(&sqs.DeleteMessageInput{
							QueueUrl:      &f.queue,
							ReceiptHandle: message.ReceiptHandle,
						}); e != nil {
							return fmt.Errorf("error delete message: %w", e)
						}
						return nil
					},

					Nack: func() error {
						if _, e := f.sqs.ChangeMessageVisibility(&sqs.ChangeMessageVisibilityInput{
							QueueUrl:          &f.queue,
							ReceiptHandle:     message.ReceiptHandle,
							VisibilityTimeout: aws.Int64(0),
						}); e != nil {
							return fmt.Errorf("error change message visibility: %w", e)
						}
						return nil
					},
				}
			}

			time.Sleep(time.Second)
		}

	}()
}
