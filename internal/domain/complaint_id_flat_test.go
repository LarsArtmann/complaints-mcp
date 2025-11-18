package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComplaintID_FlatJSON_BUG_FIX(t *testing.T) {
	t.Run("critical bug fix - flat JSON structure", func(t *testing.T) {
		id := ComplaintID("550e8400-e29b-41d4-a716-446655440000")
		
		data, err := json.Marshal(id)
		require.NoError(t, err)
		
		// Verify flat structure - this is the CRITICAL BUG FIX
		var result map[string]any
		err = json.Unmarshal(data, &result)
		require.NoError(t, err)
		
		// Should be flat string, not nested object
		idValue, exists := result["id"]
		require.True(t, exists, "id field should exist")
		
		idStr, isString := idValue.(string)
		assert.True(t, isString, "id should be flat string, got %T", idValue)
		assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", idStr)
		
		// CRITICAL: Ensure no nested "Value" objects
		assert.NotContains(t, string(data), `"Value"`, "JSON should not contain nested Value objects")
	})
	
	t.Run("unmarshal from flat JSON works", func(t *testing.T) {
		jsonData := `{"id":"550e8400-e29b-41d4-a716-446655440000"}`
		
		var id ComplaintID
		err := json.Unmarshal([]byte(jsonData), &id)
		require.NoError(t, err)
		
		assert.Equal(t, ComplaintID("550e8400-e29b-41d4-a716-446655440000"), id)
		assert.True(t, id.IsValid())
	})
	
	t.Run("reject nested JSON format", func(t *testing.T) {
		jsonData := `{"id":{"Value":"550e8400-e29b-41d4-a716-446655440000"}}`
		
		var id ComplaintID
		err := json.Unmarshal([]byte(jsonData), &id)
		assert.Error(t, err)
	})
}