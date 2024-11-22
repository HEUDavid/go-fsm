package rmq

import (
	"context"
	"fmt"
	"github.com/HEUDavid/go-fsm/pkg/mq"
	"github.com/HEUDavid/go-fsm/pkg/util"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitmqClient struct {
	conn           *amqp.Connection
	channel        *amqp.Channel
	url            string
	queueName      string
	consumerBuffer chan *mq.Message
	done           chan bool
}

func (r *RabbitmqClient) Type() string {
	if r.consumerBuffer == nil {
		return "producer"
	}
	return "consumer"
}

func (r *RabbitmqClient) Connect() error {
	var err error

	if r.conn, err = amqp.Dial(r.url); err != nil {
		return fmt.Errorf("dial Err: %v", err)
	}

	if r.channel, err = r.conn.Channel(); err != nil {
		return fmt.Errorf("channel Err: %v", err)
	}

	if _, err = r.channel.QueueDeclare(
		r.queueName,
		false,
		false,
		false,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("queueDeclare Err: %v", err)
	}

	if r.Type() == "producer" {
		return nil
	}

	return r.Consume()
}

func (r *RabbitmqClient) Consume() error {
	deliveries, err := r.channel.Consume(
		r.queueName,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	for delivery := range deliveries {
		r.consumerBuffer <- &mq.Message{
			C:    context.Background(),
			Body: string(delivery.Body),
			Ack:  func() error { return delivery.Ack(false) },
			Nack: func() error { return delivery.Nack(false, true) },
		}
	}
	return nil
}

func (r *RabbitmqClient) Reconnect() {
	for {
		select {
		case <-r.done:
			return
		default:
		}

		if err := r.Connect(); err != nil {
			log.Printf("%s(%p) connect Err: %v\n", r.Type(), r, err)
			time.Sleep(time.Second)

			continue
		}

		connClose := make(chan *amqp.Error)
		r.conn.NotifyClose(connClose)

		select {
		case <-connClose:
			log.Printf("%s(%p) reconnect...", r.Type(), r)
		case <-r.done:
			return
		}
	}
}

func (r *RabbitmqClient) Start() {
	go r.Reconnect()
	time.Sleep(time.Second)
}

func (r *RabbitmqClient) Stop() {
	r.done <- true
	if r.channel != nil {
		_ = r.channel.Close()
	}
	if r.conn != nil {
		_ = r.conn.Close()
	}
}

func (r *RabbitmqClient) Publish(body string) error {
	if r.conn == nil || r.conn.IsClosed() {
		return fmt.Errorf("bad Connection")
	}

	return r.channel.Publish(
		"",
		r.queueName,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		},
	)
}

func NewRmqClient(url, queue string, consumerBuffer chan *mq.Message) *RabbitmqClient {
	return &RabbitmqClient{
		url:            url,
		queueName:      queue,
		consumerBuffer: consumerBuffer,
		done:           make(chan bool),
	}
}

type Factory struct {
	Section  string
	buffer   chan *mq.Message
	producer *RabbitmqClient
	consumer *RabbitmqClient
}

func (f *Factory) GetMQSection() string {
	return f.Section
}

func (f *Factory) InitMQ(config util.Config) error {
	url := fmt.Sprintf(
		"amqp://%s:%s@%s:%d%s",
		config["user"], config["password"], config["host"], config["port"], config["vhost"],
	)
	queue := config["queue"].(string)

	f.consumer = NewRmqClient(url, queue, make(chan *mq.Message))
	f.buffer = f.consumer.consumerBuffer

	f.producer = NewRmqClient(url, queue, nil)
	f.producer.Start()

	return nil
}

func (f *Factory) PublishMessage(c context.Context, msg string) error {
	return f.producer.Publish(msg)
}

func (f *Factory) FetchMessage(c context.Context) mq.Message {
	msg, ok := <-f.buffer
	if !ok {
		return mq.Message{C: c}
	}
	return *msg
}

func (f *Factory) Start() {
	f.consumer.Start()
}

func (f *Factory) Stop() {
	f.producer.Stop()
	f.consumer.Stop()
}
