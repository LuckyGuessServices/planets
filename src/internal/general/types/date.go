package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/araddon/dateparse"
)

type Date struct {
	time.Time
}

func NewDate(rawTime time.Time) Date {
	y, m, d := rawTime.Date()

	return Date{
		Time: time.Date(y, m, d, 0, 0, 0, 0, rawTime.Location()),
	}
}

func (date Date) String() string {
	return date.Format(time.DateOnly)
}

func (date Date) Value() (driver.Value, error) {
	return date.String(), nil
}

//goland:noinspection GoMixedReceiverTypes
func (date *Date) Scan(value any) error {
	if valueValidated, ok := value.(time.Time); ok {
		date.Time = valueValidated

		return nil
	}

	return fmt.Errorf("unable to scan %v into Date", value)
}

func (date Date) MarshalJSON() ([]byte, error) {
	return json.Marshal(date.String())
}

//goland:noinspection GoMixedReceiverTypes
func (date *Date) UnmarshalJSON(data []byte) error {
	var valueString string
	if err := json.Unmarshal(data, &valueString); err != nil {
		return err
	}

	valueTime, errParse := dateparse.ParseAny(valueString, dateparse.PreferMonthFirst(true))
	if errParse != nil {
		return errParse
	}

	date.Time = valueTime

	return nil
}
