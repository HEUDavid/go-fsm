package main

import (
	"context"
	"fmt"
	"github.com/HEUDavid/go-fsm/pkg/mq"
	"github.com/HEUDavid/go-fsm/pkg/mq/aws"
	"github.com/HEUDavid/go-fsm/pkg/mq/rmq"
	"github.com/HEUDavid/go-fsm/pkg/util"
	"log"
	"time"
)

func genMQ(t string) mq.IMQ {
	switch t {
	case "sqs":
		return &aws.Factory{Section: "sqs_aws"}
	case "rmq":
		return &rmq.Factory{Section: "rmq_cloud"}
	}
	return nil
}

func main() {
	q := genMQ("rmq")
	conf := (*util.GetConfig())[q.GetMQSection()].(util.Config)
	_ = q.InitMQ(conf)
	q.Start()

	go func() {
		for {
			msg := q.FetchMessage(context.TODO())
			log.Printf("FetchMessage  : %s", msg.Body)
			log.Println("ACK:", msg.Ack())
		}
	}()

	go func() {
		idx := 0
		for {
			msg := fmt.Sprintf("Hello %d", idx)
			err := q.PublishMessage(context.TODO(), msg)
			if err != nil {
				log.Printf("PublishMessage Err: %v", err)
			} else {
				log.Printf("PublishMessage: %s", msg)
				idx += 1
			}
			time.Sleep(3 * time.Second)
		}
	}()

	forever := make(chan bool)
	<-forever
}
