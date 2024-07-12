package util

import (
	"log"
	"testing"
)

func TestUniqueID(t *testing.T) {
	log.Println(UniqueID())
	log.Println(UniqueID())
	log.Println(UniqueID())
}
