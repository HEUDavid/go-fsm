package rmq

import (
	"context"
	"fmt"
	"github.com/HEUDavid/go-fsm/pkg/mq"
	"github.com/HEUDavid/go-fsm/pkg/util"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Factory struct {
	Section string
	buffer  chan mq.Message
	conn    *amqp.Connection
	ch      *amqp.Channel
	q       amqp.Queue
}

func (f *Factory) GetMQSection() string {
	return f.Section
}

func (f *Factory) InitMQ(config util.Config) error {
	url := fmt.Sprintf(
		"amqp://%s:%s@%s:%d%s",
		config["user"], config["password"], config["host"], config["port"], config["vhost"],
	)

	conn, err := amqp.Dial(url)
	if err != nil {
		return err
	}

	ch, err := conn.Channel()
	if err != nil {
		return err
	}

	q, err := ch.QueueDeclare(
		config["queue"].(string),
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	f.conn = conn
	f.ch = ch
	f.q = q

	f.buffer = make(chan mq.Message)

	return nil
}

func (f *Factory) PublishMessage(c context.Context, msg string) error {
	err := f.ch.Publish(
		"",
		f.q.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(msg),
		})
	return err
}

func (f *Factory) FetchMessage(c context.Context) mq.Message {
	msg, ok := <-f.buffer
	if !ok {
		return mq.Message{C: c}
	}
	return msg
}

func (f *Factory) Start() {
	deliveries, err := f.ch.Consume(
		f.q.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		panic(err)
	}

	for delivery := range deliveries {
		f.buffer <- mq.Message{
			C:    context.Background(),
			Msg:  string(delivery.Body),
			Ack:  func() error { return delivery.Ack(false) },
			Nack: func() error { return delivery.Nack(false, true) },
		}
	}
}
