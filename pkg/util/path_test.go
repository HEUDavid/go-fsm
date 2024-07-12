package util

import (
	"log"
	"testing"
)

func TestFindProjectRoot(t *testing.T) {
	log.Println("Project Root:", FindProjectRoot())
}
