package util

import (
	"fmt"
	"testing"
)

func TestFindProjectRoot(t *testing.T) {
	fmt.Println("Project Root:", FindProjectRoot())
}
