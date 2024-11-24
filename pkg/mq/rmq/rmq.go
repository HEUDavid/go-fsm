package rmq

import (
	"context"
	"fmt"
	"github.com/HEUDavid/go-fsm/pkg/mq"
	"github.com/HEUDavid/go-fsm/pkg/util"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"time"
)

type RabbitmqClient struct {
	conn      *amqp.Connection
	channel   *amqp.Channel
	buffer    chan *mq.Message
	url       string
	queueName string
}

func NewRmqClient(url, queue string) *RabbitmqClient {
	return &RabbitmqClient{
		buffer:    make(chan *mq.Message),
		url:       url,
		queueName: queue,
	}
}

func (r *RabbitmqClient) Connect() error {
	var err error

	if r.conn, err = amqp.Dial(r.url); err != nil {
		return err
	}

	if r.channel, err = r.conn.Channel(); err != nil {
		return err
	}

	if _, err = r.channel.QueueDeclare(
		r.queueName,
		false,
		false,
		false,
		false,
		nil,
	); err != nil {
		return err
	}

	log.Printf("[FSM] rabbitmq(%p) connect success.\n", r)
	return nil
}

func (r *RabbitmqClient) Reconnect() {
	for {
		if err := r.Connect(); err != nil {
			log.Printf("[FSM] rabbitmq(%p) connect Err: %v\n", r, err)
			time.Sleep(time.Second * 2)
			continue
		}

		// 阻塞监听通道关闭事件
		connClose := make(chan *amqp.Error)
		r.conn.NotifyClose(connClose)

		select {
		case <-connClose:
			log.Printf("[FSM] rabbitmq(%p) closed, reconnect...", r)
		}
	}
}

func (r *RabbitmqClient) Start() {
	go r.Reconnect()
}

func (r *RabbitmqClient) Stop() {
	if r.channel != nil {
		_ = r.channel.Close()
	}
	if r.conn != nil {
		_ = r.conn.Close()
	}
}

func (r *RabbitmqClient) Consume() error {
	for {
		if r.channel == nil || r.channel.IsClosed() {
			time.Sleep(time.Second)
			continue
		}
		log.Println("[FSM] rabbitmq start consuming...")

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
			log.Printf("[FSM] rabbitmq consume Err: %v\n", err)
			continue
		}

		for delivery := range deliveries {
			r.buffer <- &mq.Message{
				C:    context.Background(),
				Body: string(delivery.Body),
				Ack:  func() error { return delivery.Ack(false) },
				Nack: func() error { return delivery.Nack(false, true) },
			}
		}

	}
}

func (r *RabbitmqClient) Publish(body string) error {
	if r.channel == nil || r.channel.IsClosed() {
		return fmt.Errorf("bad rabbitmq channel")
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

type Factory struct {
	Section string
	MQ      *RabbitmqClient
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

	f.MQ = NewRmqClient(url, queue)
	f.MQ.Start()
	time.Sleep(time.Second) // time for establish connection

	return nil
}

func (f *Factory) Start() {
	go f.MQ.Consume()
}

func (f *Factory) Stop() {
	f.MQ.Stop()
}

func (f *Factory) FetchMessage(c context.Context) mq.Message {
	msg, ok := <-f.MQ.buffer
	if !ok {
		return mq.Message{C: c}
	}
	return *msg
}

func (f *Factory) PublishMessage(c context.Context, msg string) error {
	return f.MQ.Publish(msg)
}
