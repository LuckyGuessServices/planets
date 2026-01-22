package luglog

import (
	"fmt"
	"log"
)

func Print(value ...any) {
	log.Print(value...)
}

func Printf(format string, value ...any) {
	Print(fmt.Sprintf(format, value...))
}

func Panic(value ...any) {
	errorMessage := fmt.Sprint(value...)
	Print(errorMessage)
	panic(errorMessage)
}

func Panicf(format string, value ...any) {
	Panic(fmt.Sprintf(format, value...))
}
