package util

import (
	"log"
	"testing"
)

func TestGetConfig(t *testing.T) {
	config := GetConfig()
	log.Printf("Config: %+v", config)
}
