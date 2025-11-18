package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComplaintID_FlatJSON_BUG_FIX(t *testing.T) {
	t.Run("CRITICAL BUG FIX - flat JSON structure", func(t *testing.T) {
		id := ComplaintID("550e8400-e29b-41d4-a716-446655440000")
		
		data, err := json.Marshal(id)
		require.NoError(t, err)
		
		// CRITICAL: Verify flat structure - the main bug fix
		// The JSON should be a flat string, not a nested object
		var flatResult string
		err = json.Unmarshal(data, &flatResult)
		require.NoError(t, err)
		
		assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", flatResult)
		
		// CRITICAL: Ensure no nested "Value" objects in the JSON
		assert.NotContains(t, string(data), `"Value"`, "JSON should not contain nested Value objects")
	})
	
	t.Run("marshal produces flat JSON string", func(t *testing.T) {
		id := ComplaintID("550e8400-e29b-41d4-a716-446655440000")
		
		data, err := json.Marshal(id)
		require.NoError(t, err)
		
		// The result should be just the string, wrapped in JSON quotes
		assert.Equal(t, `"550e8400-e29b-41d4-a716-446655440000"`, string(data))
		
		// Ensure no nesting
		assert.NotContains(t, string(data), `"Value"`)
	})
	
	t.Run("unmarshal from flat JSON string works", func(t *testing.T) {
		jsonData := `550e8400-e29b-41d4-a716-446655440000`
		
		var id ComplaintID
		err := json.Unmarshal([]byte(jsonData), &id)
		require.NoError(t, err)
		
		assert.Equal(t, ComplaintID("550e8400-e29b-41d4-a716-446655440000"), id)
		assert.True(t, id.IsValid())
	})
	
	t.Run("unmarshal from JSON object with flat ID works", func(t *testing.T) {
		jsonData := `{"id":"550e8400-e29b-41d4-a716-446655440000"}`
		
		var id ComplaintID
		err := json.Unmarshal([]byte(jsonData), &id)
		require.NoError(t, err)
		
		assert.Equal(t, ComplaintID("550e8400-e29b-41d4-a716-446655440000"), id)
		assert.True(t, id.IsValid())
	})
	
	t.Run("reject nested JSON format - old buggy structure", func(t *testing.T) {
		// This is the old buggy structure we're fixing
		jsonData := `{"id":{"Value":"550e8400-e29b-41d4-a716-446655440000"}}`
		
		var id ComplaintID
		err := json.Unmarshal([]byte(jsonData), &id)
		assert.Error(t, err)
	})
	
	t.Run("complete flat JSON structure test", func(t *testing.T) {
		id := ComplaintID("550e8400-e29b-41d4-a716-446655440000")
		
		data, err := json.Marshal(id)
		require.NoError(t, err)
		
		// Verify the complete JSON structure is flat
		type ComplaintWrapper struct {
			ID ComplaintID `json:"id"`
		}
		
		wrapper := ComplaintWrapper{ID: id}
		wrapperData, err := json.Marshal(wrapper)
		require.NoError(t, err)
		
		// The wrapper should contain flat ID
		var result map[string]any
		err = json.Unmarshal(wrapperData, &result)
		require.NoError(t, err)
		
		idValue, exists := result["id"]
		require.True(t, exists, "id field should exist")
		
		idStr, isString := idValue.(string)
		assert.True(t, isString, "id should be flat string, got %T", idValue)
		assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", idStr)
		
		// CRITICAL: Ensure no nesting in the final structure
		assert.NotContains(t, string(wrapperData), `"Value"`)
	})
}