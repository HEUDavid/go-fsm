package mysql

import (
	"context"
	"fmt"
	"github.com/HEUDavid/go-fsm/pkg/util"
	sqlDriver "github.com/go-sql-driver/mysql"
	"golang.org/x/net/proxy"
	gormDriver "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"net"
	"time"
)

type Factory struct {
	DB      *gorm.DB
	Section string
	config  util.Config
	url     string
}

func (f *Factory) GetDBSection() string {
	return f.Section
}

func (f *Factory) makeURL(net string) {
	f.url = fmt.Sprintf("%s:%s@%s(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local&tls=%t",
		f.config["user"],
		f.config["password"],
		net,
		f.config["host"],
		f.config["port"],
		f.config["dbName"],
		f.config["tls"],
	)
}

func (f *Factory) makeDsn() error {
	proxyAddr := f.config["proxyAddr"]
	if proxyAddr == nil {
		f.makeURL("tcp")
		return nil
	}

	dialer, err := proxy.SOCKS5("tcp", proxyAddr.(string), nil, proxy.Direct)
	if err != nil {
		return err
	}
	sqlDriver.RegisterDialContext("fixieDial", func(ctx context.Context, addr string) (net.Conn, error) {
		return dialer.Dial("tcp", addr)
	})
	f.makeURL("fixieDial")
	return nil
}

func (f *Factory) InitDB(config util.Config) error {
	f.config = config

	if err := f.makeDsn(); err != nil {
		return err
	}

	db, err := gorm.Open(gormDriver.Open(f.url), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("error opening database: %w", err)
	}

	// db.Debug()
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("error getting database connection pool: %w", err)
	}

	sqlDB.SetMaxOpenConns(int(f.config["maxOpenConns"].(int64)))
	sqlDB.SetMaxIdleConns(int(f.config["maxIdleConns"].(int64)))
	sqlDB.SetConnMaxLifetime(time.Hour)

	f.DB = db

	return nil
}

func (f *Factory) GetDB() *gorm.DB {
	return f.DB
}
