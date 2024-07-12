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

type Message struct {
	c   context.Context
	msg *sqs.Message
	ack func() error
}

type Factory struct {
	Section string
	buffer  chan Message
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
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(accessKey, secretKey, ""),
	})
	if err != nil {
		return err
	}

	f.sqs = sqs.New(sess)

	f.buffer = make(chan Message)

	return nil
}

func (f *Factory) PublishMessage(c context.Context, msg string) error {
	if _, err := f.sqs.SendMessage(&sqs.SendMessageInput{
		MessageBody: aws.String(msg),
		QueueUrl:    &f.queue,
	}); err != nil {
		return err
	}
	return nil
}

func (f *Factory) FetchMessage(c context.Context) (context.Context, string, mq.ACK) {
	msg, ok := <-f.buffer
	if !ok {
		return context.Background(), "", nil
	}
	return msg.c, *msg.msg.Body, msg.ack
}

func (f *Factory) Start() {
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
			f.buffer <- Message{
				c:   context.Background(),
				msg: message,
				ack: func() error {
					if _, err = f.sqs.DeleteMessage(&sqs.DeleteMessageInput{
						QueueUrl:      &f.queue,
						ReceiptHandle: message.ReceiptHandle,
					}); err != nil {
						return fmt.Errorf("error deleting message: %w", err)
					}
					return nil
				},
			}
		}

		time.Sleep(time.Second)
	}
}
