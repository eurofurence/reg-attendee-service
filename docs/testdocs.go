package docs

import "log"

// these do nothing really, but they make tests and their log output way more readable

func Given(s string) {
	log.Print(s)
}

func When(s string) {
	log.Print(s)
}

func Then(s string) {
	log.Print(s)
}

func Description(s string) {
	log.Print(s)
}

func Limitation(s string) {
	log.Print("LIMITATION: " + s)
}
