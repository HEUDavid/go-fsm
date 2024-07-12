package mq

import (
	"context"
	"github.com/HEUDavid/go-fsm/pkg/util"
)

type ACK = func() error

type IMQ interface {
	GetMQSection() string
	InitMQ(config util.Config) error
	PublishMessage(c context.Context, msg string) error
	FetchMessage(c context.Context) (context.Context, string, ACK)
	Start()
}
