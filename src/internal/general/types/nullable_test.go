package types_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/LuckyGuessServices/planets/internal/general/types"
	"github.com/LuckyGuessServices/planets/internal/test_tools"
	"github.com/stretchr/testify/require"
)

// Tests types.Nullable JSON processing.
//
// See Nullable.MarshalJSON, Nullable.UnmarshalJSON
func TestMarshallingNullable(t *testing.T) {
	tests := []struct {
		name                string
		isFieldNull         bool
		fieldValue          bool
		expectedValueInJSON []byte
	}{
		{name: "null", isFieldNull: true, fieldValue: false, expectedValueInJSON: []byte("null")},
		{name: "false", isFieldNull: false, fieldValue: false, expectedValueInJSON: []byte("false")},
		{name: "true", isFieldNull: false, fieldValue: true, expectedValueInJSON: []byte("true")},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test_tools.RunInSyncBubble(t, func(t *testing.T) {
				someStruct := &struct {
					IsActive types.Nullable[bool] `json:"is_active"`
				}{}
				if test.isFieldNull {
					someStruct.IsActive = types.NewNull[bool]()
				} else {
					someStruct.IsActive = types.NewNullable(test.fieldValue)
				}
				expectedJSON := fmt.Sprintf(`{"is_active":%s}`, test.expectedValueInJSON)

				someJSON, errMarshal := json.Marshal(&someStruct)
				require.NoError(t, errMarshal)
				require.JSONEq(t, expectedJSON, string(someJSON))

				anotherStruct := &struct {
					IsActive types.Nullable[bool] `json:"is_active"`
				}{}
				require.NoError(t, json.Unmarshal(someJSON, anotherStruct))
				if test.isFieldNull {
					require.False(t, anotherStruct.IsActive.Valid)
				} else {
					require.True(t, anotherStruct.IsActive.Valid)
					require.Equal(t, test.fieldValue, anotherStruct.IsActive.V)
				}
			})
		})
	}
}
