package internal

import (
	"github.com/HEUDavid/go-fsm/pkg/db"
	. "github.com/HEUDavid/go-fsm/pkg/metadata"
	"github.com/HEUDavid/go-fsm/pkg/mq"
	"github.com/HEUDavid/go-fsm/pkg/util"
	"gorm.io/gorm/schema"
)

type IBase[Data DataEntity] interface {
	RegisterModel(DataModel, taskModel, uniqueRequestModel schema.Tabler)
	RegisterDB(section string, db db.IDB)
	RegisterMQ(section string, mq mq.IMQ)
	RegisterFSM(fsm FSM[Data])
}

type Base[Data DataEntity] struct {
	Config *util.Config
	Models
	DBSection string
	db.IDB
	MQSection string
	mq.IMQ
	FSM[Data]
}

func (b *Base[Data]) RegisterModel(dataModel, taskModel, uniqueRequestModel schema.Tabler) {
	if !(dataModel != nil && taskModel != nil && uniqueRequestModel != nil) {
		panic("[FSM] Model data、task、unique_request should not be nil")
	}
	if !util.HasAttr(dataModel, "TaskID") {
		panic("[FSM] Model data error")
	}
	if !(util.HasAttr(taskModel, "RequestID") && util.HasAttr(taskModel, "State")) {
		panic("[FSM] Model task error")
	}
	if !(util.HasAttr(uniqueRequestModel, "RequestID") && util.HasAttr(uniqueRequestModel, "TaskID")) {
		panic("[FSM] Model unique_request error")
	}
	b.DataModel = dataModel
	b.TaskModel = taskModel
	b.UniqueRequestModel = uniqueRequestModel
}

func (b *Base[Data]) RegisterDB(section string, db db.IDB) {
	b.DBSection = section
	b.IDB = db
}

func (b *Base[Data]) RegisterMQ(section string, mq mq.IMQ) {
	b.MQSection = section
	b.IMQ = mq
}

func (b *Base[Data]) RegisterFSM(fsm FSM[Data]) {
	b.FSM = fsm
}
