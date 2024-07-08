package internal

import (
	"github.com/HEUDavid/go-fsm/pkg/db"
	. "github.com/HEUDavid/go-fsm/pkg/metadata"
	"github.com/HEUDavid/go-fsm/pkg/mq"
	"github.com/HEUDavid/go-fsm/pkg/util"
	"gorm.io/gorm/schema"
)

type IBase interface {
	RegisterFSM(fsm FSM)
	RegisterModel(extDataModel ExtDataEntity, taskModel, uniqueRequestModel schema.Tabler)
	RegisterDB(db db.IDB)
	RegisterMQ(mq mq.IMQ)
}

type Base struct {
	Config *util.Config
	FSM
	Models
	db.IDB
	mq.IMQ
}

func (b *Base) RegisterFSM(fsm FSM) {
	b.FSM = fsm
}

func (b *Base) RegisterModel(extDataModel ExtDataEntity, taskModel, uniqueRequestModel schema.Tabler) {
	if !(extDataModel != nil && taskModel != nil && uniqueRequestModel != nil) {
		panic("[FSM] Model ext_data、task、unique_request should not be nil")
	}
	if !util.HasAttr(extDataModel, "TaskID") {
		panic("[FSM] Model ext_data error")
	}
	if !(util.HasAttr(taskModel, "RequestID") && util.HasAttr(taskModel, "State")) {
		panic("[FSM] Model task error")
	}
	if !(util.HasAttr(uniqueRequestModel, "RequestID") && util.HasAttr(uniqueRequestModel, "TaskID")) {
		panic("[FSM] Model unique_request error")
	}
	b.ExtDataModel = extDataModel
	b.TaskModel = taskModel
	b.UniqueRequestModel = uniqueRequestModel
}

func (b *Base) RegisterDB(db db.IDB) {
	b.IDB = db
}

func (b *Base) RegisterMQ(mq mq.IMQ) {
	b.IMQ = mq
}
