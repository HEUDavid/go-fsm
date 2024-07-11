package db

import (
	"github.com/HEUDavid/go-fsm/pkg/util"
	"gorm.io/gorm"
)

type IDB interface {
	InitDB(config util.Config) error
	GetDBSection() string
	GetDB() *gorm.DB
}
