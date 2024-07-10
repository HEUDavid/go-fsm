package metadata

import (
	"gorm.io/gorm/schema"
	"time"
)

type Models struct {
	DataModel          schema.Tabler
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
	RequestID  string    `gorm:"unique;column:request_id;type:char(32);not null;comment:'初始请求ID'"` // 初始请求ID
	Type       string    `gorm:"column:type;type:varchar(128);not null;comment:'业务类型'"`            // 业务类型
	State      string    `gorm:"column:state;type:varchar(128);not null;comment:'任务状态'"`           // 任务状态
	Version    int       `gorm:"column:version;type:int(11);not null;default:1"`
	CreateTime time.Time `gorm:"column:create_time;type:timestamp;not null;default:current_timestamp()"`
	UpdateTime time.Time `gorm:"column:update_time;type:timestamp;not null;default:current_timestamp()"`

	Data          Data     `gorm:"-"` // Data: 自定义数据表
	SelectColumns []string `gorm:"-"` // Data: 显示声明需要更新的列，一般是zero列，其他non-zero列会自动更新不必显示声明
	OmitColumns   []string `gorm:"-"` // Data: 显示声明需要忽略更新的列
}

func (t *Task[Data]) GetData() *Data {
	return &t.Data
}

func (t *Task[Data]) SetData(Data Data) {
	t.Data = Data
}

func (t *Task[Data]) SetTaskID(taskID string) {
	t.ID = taskID
	t.Data.SetTaskID(taskID)
}

// SetSelectColumns 指定要更新的列(即使是零值)
func (t *Task[Data]) SetSelectColumns(columns []string) {
	t.SelectColumns = columns
}

// SetOmitColumns 指定要忽略的列
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
