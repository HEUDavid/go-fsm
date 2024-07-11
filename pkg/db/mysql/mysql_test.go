package mysql

import (
	"github.com/HEUDavid/go-fsm/pkg/util"
	"testing"
)

func TestInitDB(t *testing.T) {
	factory := &Factory{Section: "mysql_aws_tokyo"}
	config := (*util.GetConfig())[factory.GetDBSection()].(util.Config)
	if err := factory.InitDB(config); err != nil {
		t.Error(err)
	}
}
