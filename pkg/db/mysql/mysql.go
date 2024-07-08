package mysql

import (
	"fmt"
	"github.com/HEUDavid/go-fsm/pkg/util"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Factory struct {
	DB *gorm.DB
}

func (f *Factory) InitDB(config util.Config) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local",
		config["user"],
		config["password"],
		config["host"],
		config["port"],
		config["dbName"],
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("error opening database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("error getting database connection pool: %w", err)
	}

	sqlDB.SetMaxOpenConns(int(config["maxOpenConns"].(int64)))
	sqlDB.SetMaxIdleConns(int(config["maxIdleConns"].(int64)))

	f.DB = db

	return nil
}

func (f *Factory) GetDB() *gorm.DB {
	return f.DB
}
