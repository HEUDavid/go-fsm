package rmq

import (
	"context"
	"fmt"
	"github.com/HEUDavid/go-fsm/pkg/mq"
	"github.com/HEUDavid/go-fsm/pkg/util"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Message struct {
	c        context.Context
	delivery *amqp.Delivery
}

type Factory struct {
	buffer chan Message
	conn   *amqp.Connection
	ch     *amqp.Channel
	q      amqp.Queue
}

func (f *Factory) InitMQ(config util.Config) error {
	conn, err := amqp.Dial(fmt.Sprintf(
		"amqp://%s:%s@%s:%d/",
		config["user"], config["password"], config["host"], config["port"],
	))
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

	f.buffer = make(chan Message)

	return nil
}

func (f *Factory) PublishMessage(c context.Context, rawMessage string) error {
	err := f.ch.Publish(
		"",
		f.q.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(rawMessage),
		})
	return err
}

func (f *Factory) FetchMessage(c context.Context) (context.Context, string, mq.ACK) {
	msg, ok := <-f.buffer
	if !ok {
		return context.Background(), "", nil
	}
	return msg.c, string(msg.delivery.Body), func() error { return msg.delivery.Ack(false) }
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

	for d := range deliveries {
		f.buffer <- Message{c: context.Background(), delivery: &d}
	}

}
