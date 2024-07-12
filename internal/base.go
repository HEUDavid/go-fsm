package internal

import (
	"github.com/HEUDavid/go-fsm/pkg/db"
	. "github.com/HEUDavid/go-fsm/pkg/metadata"
	"github.com/HEUDavid/go-fsm/pkg/mq"
	"github.com/HEUDavid/go-fsm/pkg/util"
	"gorm.io/gorm/schema"
)

type IBase[Data DataEntity] interface {
	RegisterModel(dataModel, taskModel, uniqueRequestModel schema.Tabler)
	RegisterDB(db db.IDB)
	RegisterMQ(mq mq.IMQ)
	RegisterFSM(fsm FSM[Data])
	RegisterGenerator(genID func() string)
}

type Base[Data DataEntity] struct {
	Config *util.Config
	Models
	db.IDB
	mq.IMQ
	FSM[Data]
	GenID func() string // ID Generator
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

func (b *Base[Data]) RegisterDB(db db.IDB) {
	b.IDB = db
}

func (b *Base[Data]) RegisterMQ(mq mq.IMQ) {
	b.IMQ = mq
}

func (b *Base[Data]) RegisterFSM(fsm FSM[Data]) {
	b.FSM = fsm
}

func (b *Base[Data]) RegisterGenerator(genID func() string) {
	b.GenID = genID
}
