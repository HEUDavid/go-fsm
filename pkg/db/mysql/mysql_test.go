package mysql

import (
	"github.com/HEUDavid/go-fsm/pkg/util"
	"testing"
)

func TestInitDB(t *testing.T) {
	config := (*util.GetConfig())["mysql_aiven"].(util.Config)
	factory := &Factory{}
	if err := factory.InitDB(config); err != nil {
		t.Error(err)
	}
}
