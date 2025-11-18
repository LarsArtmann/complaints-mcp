package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCriticalJSONBugFix(t *testing.T) {
	t.Run("✅ OLD BUG - nested JSON structure", func(t *testing.T) {
		// This was the buggy structure that needs to be fixed
		oldBuggyJSON := `{"id":{"Value":"550e8400-e29b-41d4-a716-446655440000"}}`
		
		var newID ComplaintID
		err := json.Unmarshal([]byte(oldBuggyJSON), &newID)
		
		// ✅ GOOD: Should reject the old buggy nested format
		assert.Error(t, err, "Should reject old nested format")
	})
	
	t.Run("✅ NEW FIX - flat JSON structure", func(t *testing.T) {
		id := ComplaintID("550e8400-e29b-41d4-a716-446655440000")
		
		// Serialize to JSON
		data, err := json.Marshal(id)
		require.NoError(t, err)
		
		// ✅ CRITICAL FIX: Should produce flat string, not nested object
		jsonStr := string(data)
		assert.Equal(t, `"550e8400-e29b-41d4-a716-446655440000"`, jsonStr)
		
		// ✅ CRITICAL FIX: Should not contain nested "Value" objects
		assert.NotContains(t, jsonStr, `"Value"`, "Should not have nested Value objects")
	})
	
	t.Run("✅ NEW FIX - complete complaint with flat IDs", func(t *testing.T) {
		complaint := &Complaint{
			ID:             ComplaintID("550e8400-e29b-41d4-a716-446655440000"),
			AgentID:        AgentID("AI-Assistant"),
			SessionID:      SessionID("dev-session"),
			ProjectID:      ProjectID("my-project"),
			TaskDescription: "Test complaint description",
			Severity:       SeverityMedium,
			Timestamp:      testTimestamp,
			ResolutionState: ResolutionStateOpen,
		}
		
		// Serialize to JSON
		data, err := json.Marshal(complaint)
		require.NoError(t, err)
		
		jsonStr := string(data)
		
		// ✅ CRITICAL FIX: All IDs should be flat strings
		assert.Contains(t, jsonStr, `"id":"550e8400-e29b-41d4-a716-446655440000"`)
		assert.Contains(t, jsonStr, `"agent_id":"AI-Assistant"`)
		assert.Contains(t, jsonStr, `"session_id":"dev-session"`)
		assert.Contains(t, jsonStr, `"project_id":"my-project"`)
		
		// ✅ CRITICAL FIX: Should not contain any nested "Value" objects
		assert.NotContains(t, jsonStr, `"Value"`, "Should not have nested Value objects")
	})
	
	t.Run("✅ NEW FIX - roundtrip with flat IDs", func(t *testing.T) {
		originalComplaint := &Complaint{
			ID:             ComplaintID("550e8400-e29b-41d4-a716-446655440000"),
			AgentID:        AgentID("AI-Assistant"),
			SessionID:      SessionID("dev-session"),
			ProjectID:      ProjectID("my-project"),
			TaskDescription: "Test complaint description",
			Severity:       SeverityMedium,
			Timestamp:      testTimestamp,
			ResolutionState: ResolutionStateOpen,
		}
		
		// Serialize to JSON
		data, err := json.Marshal(originalComplaint)
		require.NoError(t, err)
		
		// Deserialize back
		var restoredComplaint Complaint
		err = json.Unmarshal(data, &restoredComplaint)
		require.NoError(t, err)
		
		// ✅ All phantom types should be preserved
		assert.Equal(t, originalComplaint.ID, restoredComplaint.ID)
		assert.Equal(t, originalComplaint.AgentID, restoredComplaint.AgentID)
		assert.Equal(t, originalComplaint.SessionID, restoredComplaint.SessionID)
		assert.Equal(t, originalComplaint.ProjectID, restoredComplaint.ProjectID)
		
		// ✅ All should still be valid
		assert.True(t, restoredComplaint.ID.IsValid())
		assert.True(t, restoredComplaint.AgentID.IsValid())
		assert.True(t, restoredComplaint.SessionID.IsValid())
		assert.True(t, restoredComplaint.ProjectID.IsValid())
	})
}

// Test helper
var testTimestamp = time.Date(2024, 11, 9, 12, 18, 30, 0, time.UTC)