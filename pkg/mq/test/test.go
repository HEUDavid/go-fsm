package main

import (
	"context"
	"fmt"
	"github.com/HEUDavid/go-fsm/pkg/mq"
	"github.com/HEUDavid/go-fsm/pkg/mq/aws"
	"github.com/HEUDavid/go-fsm/pkg/mq/rmq"
	"github.com/HEUDavid/go-fsm/pkg/util"
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
	_mq := genMQ("sqs")

	config := (*util.GetConfig())[_mq.GetMQSection()].(util.Config)
	if err := _mq.InitMQ(config); err != nil {
		panic(err)
	}

	go _mq.Start()

	go func() {
		for {
			_, msg, ack := _mq.FetchMessage(context.TODO())
			fmt.Println("FetchMessage:", msg)
			if ack != nil {
				_ = ack()
			}
		}
	}()

	go func() {
		for i := 1; ; i++ {
			msg := fmt.Sprintf("Hello %d", i)
			_ = _mq.PublishMessage(context.TODO(), msg)
			fmt.Println("PublishMessage:", msg)

			time.Sleep(3 * time.Second)
		}
	}()

	forever := make(chan bool)
	fmt.Println("Exit press CTRL+C...")
	<-forever

}
