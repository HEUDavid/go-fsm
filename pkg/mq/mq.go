package mq

import (
	"context"
	"github.com/HEUDavid/go-fsm/pkg/util"
)

type Message struct {
	C    context.Context
	Msg  string
	Ack  func() error
	Nack func() error
}

type IMQ interface {
	GetMQSection() string
	InitMQ(config util.Config) error
	PublishMessage(c context.Context, msg string) error
	FetchMessage(c context.Context) Message
	Start()
}
