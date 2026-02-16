package types

import (
	"database/sql"
	"encoding/json"
)

// Nullable is a generic wrapper for nullable database fields.
type Nullable[Type any] struct {
	sql.Null[Type]
}

// NewNull provides a Nullable struct with no (nil) value. To create with a value, call NewNullable instead.
func NewNull[Type any]() Nullable[Type] {
	return Nullable[Type]{}
}

// NewNullable provides a Nullable struct with an actual value. To specify `nil` explicitly call NewNull instead.
func NewNullable[Type any](value Type) Nullable[Type] {
	return Nullable[Type]{sql.Null[Type]{V: value, Valid: true}}
}

func (nullable Nullable[Type]) MarshalJSON() ([]byte, error) {
	if !nullable.Valid {
		return []byte("null"), nil
	}

	return json.Marshal(nullable.V)
}

//goland:noinspection GoMixedReceiverTypes
func (nullable *Nullable[Type]) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		nullable.V, nullable.Valid = *new(Type), false

		return nil
	}

	if err := json.Unmarshal(data, &nullable.V); err != nil {
		return err
	}
	nullable.Valid = true

	return nil
}
