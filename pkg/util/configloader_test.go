package util

import (
	"fmt"
	"testing"
)

func TestGetConfig(t *testing.T) {
	config := GetConfig()
	fmt.Printf("Config: %+v\n", config)
}
