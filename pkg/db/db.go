package db

import (
	"github.com/HEUDavid/go-fsm/pkg/util"
	"gorm.io/gorm"
)

type IDB interface {
	GetDBSection() string
	InitDB(config util.Config) error
	GetDB() *gorm.DB
}
