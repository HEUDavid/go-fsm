package util

import (
	"log"
	"testing"
)

func TestGenID(t *testing.T) {
	log.Println(GenID())
	log.Println(GenID())
	log.Println(GenID())
}
