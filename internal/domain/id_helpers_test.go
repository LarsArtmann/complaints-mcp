package domain

import (
	"encoding/json"
	"testing"

	"github.com/larsartmann/go-composable-business-types/id"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func AssertFlatJSONMarshaling[Brand any](
	tb testing.TB,
	brandedID id.ID[Brand, string],
	expectedValue string,
) {
	tb.Helper()

	data, err := json.Marshal(brandedID)
	require.NoError(tb, err)

	var flatResult string
	err = json.Unmarshal(data, &flatResult)
	require.NoError(tb, err)
	assert.Equal(tb, expectedValue, flatResult)

	assert.NotContains(tb, string(data), `"Value"`)
}

func AssertUnmarshalFromFlatJSON[Brand any](
	tb testing.TB,
	brandedID id.ID[Brand, string],
	jsonValue string,
) {
	tb.Helper()

	data := []byte(jsonValue)
	var result id.ID[Brand, string]
	err := json.Unmarshal(data, &result)
	require.NoError(tb, err)
	assert.Equal(tb, brandedID, result)
	assert.Equal(tb, false, result.IsZero())
}

func AssertRejectNestedJSONFormat[Brand any](tb testing.TB, jsonData string) {
	tb.Helper()

	var result id.ID[Brand, string]
	err := json.Unmarshal([]byte(jsonData), &result)
	assert.Error(tb, err)
}
