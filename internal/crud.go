package internal

import (
	"context"
	"fmt"
	. "github.com/HEUDavid/go-fsm/pkg/metadata"
	"github.com/HEUDavid/go-fsm/pkg/util"
	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func AddTaskFlow[ExtData ExtDataEntity](tx *gorm.DB, m Models, task *Task[ExtData], queryFirst bool) error {
	if m.TaskFlowModel == nil || m.ExtDataModel == nil {
		return nil
	}
	return nil
}

func AddUnique[ExtData ExtDataEntity](tx *gorm.DB, uniqueRequestModel schema.Tabler, task *Task[ExtData], needModifyTaskID bool) (bool, error) {
	uniqueReq := struct {
		RequestID string
		TaskID    string
	}{
		task.RequestID,
		task.ID,
	}

	err := tx.Table(uniqueRequestModel.TableName()).Create(&uniqueReq).Error

	if err != nil {
		mysqlErr, ok := err.(*mysql.MySQLError)
		if !ok {
			return false, err
		}
		switch mysqlErr.Number {
		case 1062:
			if needModifyTaskID { // Use the TaskID recorded in the DB to assign values, making the interface idempotent.
				if err = tx.Table(uniqueRequestModel.TableName()).Where("request_id = ?", task.RequestID).Scan(&uniqueReq).Error; err != nil {
					return true, err
				}
				task.SetTaskID(uniqueReq.TaskID)
			}
			return true, nil
		}
		return false, err
	} else {
		return false, nil
	}
}

func CreateTaskTx[ExtData ExtDataEntity](c context.Context, db *gorm.DB, m Models, task *Task[ExtData]) error {
	if err := db.Transaction(func(tx *gorm.DB) error { return _createTaskTx(c, tx, m, task) }); err != nil {
		return err
	}
	return nil
}

func _createTaskTx[ExtData ExtDataEntity](c context.Context, tx *gorm.DB, m Models, task *Task[ExtData]) error {
	keyConflict, e := AddUnique(tx, m.UniqueRequestModel, task, true)
	if e != nil {
		return e
	}
	if keyConflict {
		return nil
	}

	if e = tx.Table(m.TaskModel.TableName()).Create(&task).Error; e != nil {
		return e
	}
	if e = tx.Table(m.ExtDataModel.TableName()).Create(task.GetExtData()).Error; e != nil {
		return e
	}

	if e = AddTaskFlow(tx, m, task, false); e != nil {
		return e
	}

	return nil
}

func QueryTaskTx[ExtData ExtDataEntity](c context.Context, db *gorm.DB, taskModel, extDataModel schema.Tabler, task *Task[ExtData]) error {
	stmt := db
	if task.ID != "" {
		// When the destination object has a primary key value, it will be used to construct the condition
	} else {
		stmt = stmt.Where("request_id = ?", task.RequestID)
	}

	if err := stmt.Table(taskModel.TableName()).First(task).Error; err != nil {
		return err
	}
	if extDataModel != nil {
		if err := db.Table(extDataModel.TableName()).Where("task_id = ?", task.ID).Find(task.ExtData).Error; err != nil {
			return err
		}
	}

	return nil
}

func QueryTaskState(c context.Context, tx *gorm.DB, taskModel schema.Tabler, taskID string) (*string, error) {
	var state string
	if err := tx.Table(taskModel.TableName()).Select("state").Where("id = ?", taskID).Scan(&state).Error; err != nil {
		return nil, err
	}
	return &state, nil
}

func UpdateTaskTx[ExtData ExtDataEntity](c context.Context, db *gorm.DB, m Models, task *Task[ExtData], fsm FSM) error {
	if err := db.Transaction(func(tx *gorm.DB) error { return _updateTaskTx(c, tx, m, task, fsm) }); err != nil {
		return err
	}
	return nil
}

func _updateTaskTx[ExtData ExtDataEntity](c context.Context, tx *gorm.DB, m Models, task *Task[ExtData], fsm FSM) error {
	keyConflict, e := AddUnique(tx, m.UniqueRequestModel, task, false)
	if e != nil {
		return e
	}
	if keyConflict {
		return nil
	}

	var currentTask Task[ExtDataEntity]
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

	if e = UpdateExtData(c, tx, m.ExtDataModel, task); e != nil {
		return e
	}

	if e = AddTaskFlow(tx, m, task, true); e != nil {
		return e
	}

	return nil
}

func UpdateExtData[ExtData ExtDataEntity](c context.Context, tx *gorm.DB, extDataModel schema.Tabler, task *Task[ExtData]) error {
	if task.GetExtData() == nil {
		return nil
	}

	_query := func(_tx *gorm.DB) *gorm.DB {
		return _tx.Table(extDataModel.TableName()).Where("task_id = ?", task.ID)
	}

	if len(task.SelectColumns) <= 0 {
		if err := _query(tx).Omit(task.OmitColumns...).Updates(task.ExtData).Error; err != nil {
			return err
		}
		return nil
	}

	sqlStr1 := tx.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return _query(tx).Omit(task.OmitColumns...).Updates(task.ExtData)
	})
	sqlStr2 := tx.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return _query(tx).Select(task.SelectColumns).Updates(task.ExtData)
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
