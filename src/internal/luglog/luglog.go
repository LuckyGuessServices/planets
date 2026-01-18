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

func Fatal(value ...any) {
	errorMessage := fmt.Sprint(value...)
	Print(errorMessage)
	panic(errorMessage)
}

func Fatalf(format string, value ...any) {
	Fatal(fmt.Sprintf(format, value...))
}
