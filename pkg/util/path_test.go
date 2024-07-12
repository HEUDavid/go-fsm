package util

import (
	"log"
	"testing"
)

func TestFindProjectRoot(t *testing.T) {
	log.Printf("Project Root: %s", FindProjectRoot())
}
