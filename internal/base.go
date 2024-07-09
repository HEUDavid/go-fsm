package internal

import (
	"github.com/HEUDavid/go-fsm/pkg/db"
	. "github.com/HEUDavid/go-fsm/pkg/metadata"
	"github.com/HEUDavid/go-fsm/pkg/mq"
	"github.com/HEUDavid/go-fsm/pkg/util"
	"gorm.io/gorm/schema"
)

type IBase[ExtData ExtDataEntity] interface {
	RegisterModel(extDataModel, taskModel, uniqueRequestModel schema.Tabler)
	RegisterDB(db db.IDB)
	RegisterMQ(mq mq.IMQ)
	RegisterFSM(fsm FSM[ExtData])
}

type Base[ExtData ExtDataEntity] struct {
	Config *util.Config
	Models
	db.IDB
	mq.IMQ
	FSM[ExtData]
}

func (b *Base[ExtData]) RegisterModel(extDataModel, taskModel, uniqueRequestModel schema.Tabler) {
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

func (b *Base[ExtData]) RegisterDB(db db.IDB) {
	b.IDB = db
}

func (b *Base[ExtData]) RegisterMQ(mq mq.IMQ) {
	b.IMQ = mq
}

func (b *Base[ExtData]) RegisterFSM(fsm FSM[ExtData]) {
	b.FSM = fsm
}
