package main

import (
	"context"
	"fmt"
	"github.com/HEUDavid/go-fsm/pkg/mq/rmq"
	"github.com/HEUDavid/go-fsm/pkg/util"
	"time"
)

func main() {
	config := (*util.GetConfig())["rmq"].(util.Config)

	mq := rmq.Factory{}
	if err := mq.InitMQ(config); err != nil {
		panic(err)
	}

	go mq.Start()

	go func() {
		for {
			_, msg, ack := mq.FetchMessage(context.TODO())
			fmt.Println("FetchMessage:", msg)
			if ack != nil {
				_ = ack()
			}
		}
	}()

	go func() {
		for i := 1; ; i++ {
			msg := fmt.Sprintf("Hello %d", i)
			_ = mq.PublishMessage(context.TODO(), msg)
			fmt.Println("PublishMessage:", msg)

			time.Sleep(3 * time.Second)
		}
	}()

	forever := make(chan bool)
	fmt.Println("Exit press CTRL+C...")
	<-forever

}
