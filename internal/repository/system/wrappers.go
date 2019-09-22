package system

import (
	"log"
	"os"
)

var TestingMode = false
var TestingExitCounter = 0

func Exit(code int) {
	if (TestingMode) {
		TestingExitCounter++
		log.Printf("would now os.exit with code %v", code)
	} else {
		os.Exit(code)
	}
}
