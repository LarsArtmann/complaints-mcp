package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComplaintID_JSONSerialization_FlatStructure(t *testing.T) {
	t.Run("marshal produces flat JSON", func(t *testing.T) {
		id := ComplaintID("550e8400-e29b-41d4-a716-446655440000")

		data, err := json.Marshal(id)
		require.NoError(t, err)

		// Should be flat string, not nested object
		var flatResult string
		err = json.Unmarshal(data, &flatResult)
		require.NoError(t, err)
		assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", flatResult)

		// Verify no nested "Value" objects
		assert.NotContains(t, string(data), `"Value"`)
	})

	t.Run("unmarshal from flat JSON", func(t *testing.T) {
		jsonData := `"550e8400-e29b-41d4-a716-446655440000"`

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

func TestComplaintID_NewAndParse(t *testing.T) {
	t.Run("new generates valid ID", func(t *testing.T) {
		id, err := NewComplaintID()
		require.NoError(t, err)

		assert.False(t, id.IsEmpty())
		assert.True(t, id.IsValid())
		assert.NotEmpty(t, id.String())

		// Verify UUID v4 format
		assert.Regexp(t, `^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`, id.String())
	})

	t.Run("parse validates input", func(t *testing.T) {
		tests := []struct {
			name       string
			input      string
			wantErr    bool
			expectedID ComplaintID
		}{
			{"valid UUID", "550e8400-e29b-41d4-a716-446655440000", false, ComplaintID("550e8400-e29b-41d4-a716-446655440000")},
			{"valid lowercase", "9cb3bb9e-b6dc-4e02-9767-e396a42b63a6", false, ComplaintID("9cb3bb9e-b6dc-4e02-9767-e396a42b63a6")},
			{"empty string", "", true, ComplaintID("")},
			{"invalid format", "not-a-uuid", true, ComplaintID("")},
			{"wrong version", "550e8400-e29b-11d4-a716-446655440000", true, ComplaintID("")},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				id, err := ParseComplaintID(tt.input)

				if tt.wantErr {
					assert.Error(t, err)
					assert.True(t, id.IsEmpty())
					assert.False(t, id.IsValid())
				} else {
					assert.NoError(t, err)
					assert.Equal(t, tt.expectedID, id)
					assert.False(t, id.IsEmpty())
					assert.True(t, id.IsValid())
					assert.Equal(t, tt.input, id.String())
				}
			})
		}
	})
}
