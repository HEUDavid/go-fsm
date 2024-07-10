package util

import (
	"fmt"
	"testing"
)

func TestGenID(t *testing.T) {
	fmt.Println(GenID())
	fmt.Println(GenID())
	fmt.Println(GenID())
}
