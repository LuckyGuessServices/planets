package helper

import (
	"github.com/LuckyGuessServices/planets/internal/general/types"
	"github.com/LuckyGuessServices/planets/internal/luglog"
)

func NewDateParsedOrPanic(dateString string) types.Date {
	dateParsed, errParse := types.NewDateParsed(dateString)
	if errParse != nil {
		luglog.Panicf("Failed to parse '%s' as types.Date: %v", dateString, errParse)
	}

	return dateParsed
}
