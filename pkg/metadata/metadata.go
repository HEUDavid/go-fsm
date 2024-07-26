package metadata

import (
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"time"
)

type Models struct {
	DataModel          DataEntity
	TaskModel          schema.Tabler
	UniqueRequestModel schema.Tabler
	TaskFlowModel      schema.Tabler // TaskFlow and DataFlow is optional, if not set,
	DataFlowModel      schema.Tabler // no flow records will be kept
}

// DataEntity is data table which has all business columns
type DataEntity interface {
	schema.Tabler
	SetTaskID(string)
}

// Task is the main table, maintaining state, driven execution
type Task[Data DataEntity] struct {
	ID         string    `gorm:"primaryKey;column:id;type:char(32);not null"`
	RequestID  string    `gorm:"unique;column:request_id;type:char(32);not null;comment:'初始请求ID'"`       // 初始请求ID
	Type       string    `gorm:"column:type;type:varchar(128);not null;comment:'业务类型'"`                  // 业务类型
	State      string    `gorm:"index:idx_state;column:state;type:varchar(128);not null;comment:'任务状态'"` // 任务状态
	Version    uint      `gorm:"column:version;type:int unsigned;not null;default:1"`
	CreateTime time.Time `gorm:"column:create_time;type:timestamp;not null;default:CURRENT_TIMESTAMP"`
	UpdateTime time.Time `gorm:"column:update_time;type:timestamp;not null;default:CURRENT_TIMESTAMP"`

	Data          Data     `gorm:"-"`          // Data: Customized Data Tables
	SelectColumns []string `gorm:"-" json:"-"` // Data: Columns to update, including zero values
	OmitColumns   []string `gorm:"-" json:"-"` // Data: Columns to be ignored
	WithDB        *gorm.DB `gorm:"-" json:"-"`
}

func (t *Task[Data]) GetData() *Data {
	return &t.Data
}

func (t *Task[Data]) SetData(data Data) {
	t.Data = data
}

func (t *Task[Data]) SetTaskID(taskID string) {
	t.ID = taskID
	t.Data.SetTaskID(taskID)
}

// SetSelectColumns Specify columns to update (even if it is zero-valued)
func (t *Task[Data]) SetSelectColumns(columns []string) {
	t.SelectColumns = columns
}

// SetOmitColumns Specify columns to ignore
func (t *Task[Data]) SetOmitColumns(columns []string) {
	t.OmitColumns = columns
}

func GenTaskInstance[Data DataEntity](requestID, taskID string, data Data) *Task[Data] {
	task := &Task[Data]{Data: data}
	task.RequestID = requestID
	if taskID != "" {
		task.SetTaskID(taskID)
	}
	return task
}
