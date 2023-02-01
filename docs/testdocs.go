package docs

import (
	"fmt"
)

// these do nothing really, but they make tests and their log output way more readable

const indentation = "             "
const iLimitation = "LIMITATION:  "

func Given(s string) {
	fmt.Println(indentation, s)
}

func When(s string) {
	fmt.Println(indentation, s)
}

func Then(s string) {
	fmt.Println(indentation, s)
}

func Description(s string) {
	fmt.Println(indentation, s)
}

func Limitation(s string) {
	fmt.Println(iLimitation + s)
}
