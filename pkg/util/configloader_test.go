package util

import (
	"log"
	"testing"
)

func TestGetConfig(t *testing.T) {
	_conf := GetConfig()
	log.Printf("Config: %+v", _conf)
}
