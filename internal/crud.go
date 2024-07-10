package internal

import (
	"context"
	"fmt"
	. "github.com/HEUDavid/go-fsm/pkg/metadata"
	"github.com/HEUDavid/go-fsm/pkg/util"
	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

func AddTaskFlow[Data DataEntity](tx *gorm.DB, m Models, task *Task[Data], queryFirst bool) error {
	if m.TaskFlowModel == nil || m.DataModel == nil {
		return nil
	}
	return nil
}

func AddUnique[Data DataEntity](tx *gorm.DB, m Models, task *Task[Data], needModifyTaskID bool) (bool, error) {
	uniqueReq := struct {
		RequestID string
		TaskID    string
	}{
		task.RequestID,
		task.ID,
	}

	err := tx.Table(m.UniqueRequestModel.TableName()).Create(&uniqueReq).Error
	if err == nil {
		return false, nil
	}

	mysqlErr, ok := err.(*mysql.MySQLError)
	if !ok {
		return false, err
	}
	switch mysqlErr.Number {
	case 1062:
		if needModifyTaskID { // Use the TaskID recorded in the DB to assign values, making the interface idempotent.
			if err = tx.Table(m.UniqueRequestModel.TableName()).Where("request_id = ?", task.RequestID).Scan(&uniqueReq).Error; err != nil {
				return true, err
			}
			task.SetTaskID(uniqueReq.TaskID)
		}
		return true, nil
	}
	return false, err
}

func CreateTaskTx[Data DataEntity](c context.Context, db *gorm.DB, m Models, task *Task[Data]) error {
	if err := db.Transaction(func(tx *gorm.DB) error { return _createTaskTx(c, tx, m, task) }); err != nil {
		return err
	}
	return nil
}

func _createTaskTx[Data DataEntity](c context.Context, tx *gorm.DB, m Models, task *Task[Data]) error {
	keyConflict, e := AddUnique(tx, m, task, true)
	if e != nil {
		return e
	}
	if keyConflict {
		return nil
	}

	if e = tx.Table(m.TaskModel.TableName()).Create(&task).Error; e != nil {
		return e
	}
	if e = tx.Table(m.DataModel.TableName()).Create(task.GetData()).Error; e != nil {
		return e
	}

	if e = AddTaskFlow(tx, m, task, false); e != nil {
		return e
	}

	return nil
}

func QueryTaskTx[Data DataEntity](c context.Context, db *gorm.DB, m Models, task *Task[Data]) error {
	_queryTask := func(_tx *gorm.DB) *gorm.DB {
		q := _tx.Table(m.TaskModel.TableName())
		if task.ID != "" {
			// When the destination object has a primary key value, it will be used to construct the condition
		} else {
			q = q.Where("request_id = ?", task.RequestID)
		}
		return q
	}
	if err := _queryTask(db).First(task).Error; err != nil {
		return err
	}

	_queryData := func(_tx *gorm.DB) *gorm.DB {
		return _tx.Table(m.DataModel.TableName()).Where("task_id = ?", task.ID)
	}
	if err := _queryData(db).Find(task.GetData()).Error; err != nil {
		return err
	}

	return nil
}

func QueryTaskState(c context.Context, tx *gorm.DB, m Models, taskID string) (*string, error) {
	var state string
	if err := tx.Table(m.TaskModel.TableName()).Select("state").Where("id = ?", taskID).Scan(&state).Error; err != nil {
		return nil, err
	}
	return &state, nil
}

func UpdateTaskTx[Data DataEntity](c context.Context, db *gorm.DB, m Models, task *Task[Data], fsm FSM[Data]) error {
	if err := db.Transaction(func(tx *gorm.DB) error { return _updateTaskTx(c, tx, m, task, fsm) }); err != nil {
		return err
	}
	return nil
}

func _updateTaskTx[Data DataEntity](c context.Context, tx *gorm.DB, m Models, task *Task[Data], fsm FSM[Data]) error {
	keyConflict, e := AddUnique(tx, m, task, false)
	if e != nil {
		return e
	}
	if keyConflict {
		return nil
	}

	var currentTask Task[Data]
	currentTask.ID = task.ID
	if e = tx.Table(m.TaskModel.TableName()).First(&currentTask).Error; e != nil {
		return e
	}
	if currentTask.Version != task.Version {
		return fmt.Errorf("task.Version not match: %d, %d", currentTask.Version, task.Version)
	}
	task.Version = currentTask.Version + 1

	_, exist := fsm.GetTransition(currentTask.State, task.State)
	if !exist {
		return fmt.Errorf("cannot transition, %s->%s", currentTask.State, task.State)
	}

	result := tx.Table(m.TaskModel.TableName()).Omit("request_id").Where("id = ? and version = ?", task.ID, currentTask.Version).Updates(task)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected <= 0 {
		return nil
	}

	if e = UpdateData(c, tx, m, task); e != nil {
		return e
	}

	if e = AddTaskFlow(tx, m, task, true); e != nil {
		return e
	}

	return nil
}

func UpdateData[Data DataEntity](c context.Context, tx *gorm.DB, m Models, task *Task[Data]) error {
	_query := func(_tx *gorm.DB) *gorm.DB {
		return _tx.Table(m.DataModel.TableName()).Where("task_id = ?", task.ID)
	}

	if len(task.SelectColumns) <= 0 {
		if err := _query(tx).Omit(task.OmitColumns...).Updates(task.GetData()).Error; err != nil {
			return err
		}
		return nil
	}

	sqlStr1 := tx.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return _query(tx).Omit(task.OmitColumns...).Updates(task.GetData())
	})
	sqlStr2 := tx.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return _query(tx).Select(task.SelectColumns).Updates(task.GetData())
	})

	sql, err := util.MergeUpdateSQL(sqlStr1, sqlStr2)
	if err != nil {
		return err
	}

	if err = tx.Exec(sql).Error; err != nil {
		return err
	}

	return nil
}
