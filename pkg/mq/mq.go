package mq

import (
	"context"
	"github.com/HEUDavid/go-fsm/pkg/util"
)

type ACK = func() error

type IMQ interface {
	InitMQ(config util.Config) error
	PublishMessage(c context.Context, msg string) error
	FetchMessage(c context.Context) (context.Context, string, ACK)
	Start()
}
