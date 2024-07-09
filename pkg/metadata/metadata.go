package metadata

import (
	"gorm.io/gorm/schema"
	"time"
)

type Models struct {
	ExtDataModel       schema.Tabler
	TaskModel          schema.Tabler
	UniqueRequestModel schema.Tabler
	TaskFlowModel      schema.Tabler // TaskFlow and ExtDataFlow is optional, if not set,
	ExtDataFlowModel   schema.Tabler // no flow records will be kept
}

// ExtDataEntity is ext_data table which has all business columns
type ExtDataEntity interface {
	schema.Tabler
	SetTaskID(string)
}

// Task is the main table, maintaining state, driven execution
type Task[ExtData ExtDataEntity] struct {
	ID         string    `gorm:"primaryKey;column:id;type:char(32);not null"`
	RequestID  string    `gorm:"unique;column:request_id;type:char(32);not null;comment:'初始请求ID'"` // 初始请求ID
	Type       string    `gorm:"column:type;type:varchar(128);not null;comment:'业务类型'"`            // 业务类型
	State      string    `gorm:"column:state;type:varchar(128);not null;comment:'任务状态'"`           // 任务状态
	Version    int       `gorm:"column:version;type:int(11);not null;default:1"`
	CreateTime time.Time `gorm:"column:create_time;type:timestamp;not null;default:current_timestamp()"`
	UpdateTime time.Time `gorm:"column:update_time;type:timestamp;not null;default:current_timestamp()"`

	ExtData       ExtData  `gorm:"-"`
	SelectColumns []string `gorm:"-"` // ExtData: 显示声明需要更新的列，一般是zero列，其他non-zero列会自动更新不必声明
	OmitColumns   []string `gorm:"-"` // ExtData: 显示申明需要忽略更新的列
}

func (t *Task[ExtData]) GetExtData() *ExtData {
	return &t.ExtData
}

func (t *Task[ExtData]) SetExtData(extData ExtData) {
	t.ExtData = extData
}

func (t *Task[ExtData]) SetTaskID(taskID string) {
	t.ID = taskID
	t.ExtData.SetTaskID(taskID)
}

// SetSelectColumns 指定要更新的列(即使是零值)
func (t *Task[ExtData]) SetSelectColumns(columns []string) {
	t.SelectColumns = columns
}

// SetOmitColumns 指定要忽略的列
func (t *Task[ExtData]) SetOmitColumns(columns []string) {
	t.OmitColumns = columns
}

func GenTaskInstance[ExtData ExtDataEntity](requestID, taskID string, extData ExtData) *Task[ExtData] {
	task := &Task[ExtData]{ExtData: extData}
	task.RequestID = requestID
	if taskID != "" {
		task.SetTaskID(taskID)
	}
	return task
}
