package domain

import (
	"encoding/json"
	"testing"

	"github.com/larsartmann/go-composable-business-types/id"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComplaintID_JSONSerialization_FlatStructure(t *testing.T) {
	t.Run("marshal produces flat JSON", func(t *testing.T) {
		complaintID := id.NewID[ComplaintBrand]("550e8400-e29b-41d4-a716-446655440000")

		data, err := json.Marshal(complaintID)
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

		var complaintID ComplaintID
		err := json.Unmarshal([]byte(jsonData), &complaintID)
		require.NoError(t, err)

		assert.Equal(
			t,
			id.NewID[ComplaintBrand]("550e8400-e29b-41d4-a716-446655440000"),
			complaintID,
		)
		assert.False(t, complaintID.IsZero())
	})

	t.Run("reject nested JSON format", func(t *testing.T) {
		jsonData := `{"id":{"Value":"550e8400-e29b-41d4-a716-446655440000"}}`

		var complaintID ComplaintID
		err := json.Unmarshal([]byte(jsonData), &complaintID)
		assert.Error(t, err)
	})
}

func TestComplaintID_NewAndParse(t *testing.T) {
	t.Run("new generates valid ID", func(t *testing.T) {
		complaintID, err := NewComplaintID()
		require.NoError(t, err)

		assert.False(t, complaintID.IsZero())
		assert.NotEmpty(t, complaintID.String())

		// Verify UUID v4 format
		assert.Regexp(
			t,
			`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`,
			complaintID.String(),
		)
	})

	t.Run("parse validates input", func(t *testing.T) {
		tests := []struct {
			name       string
			input      string
			wantErr    bool
			expectedID ComplaintID
		}{
			{
				"valid UUID",
				"550e8400-e29b-41d4-a716-446655440000",
				false,
				id.NewID[ComplaintBrand]("550e8400-e29b-41d4-a716-446655440000"),
			},
			{
				"valid lowercase",
				"9cb3bb9e-b6dc-4e02-9767-e396a42b63a6",
				false,
				id.NewID[ComplaintBrand]("9cb3bb9e-b6dc-4e02-9767-e396a42b63a6"),
			},
			{"empty string", "", true, id.NewID[ComplaintBrand]("")},
			{"invalid format", "not-a-uuid", true, id.NewID[ComplaintBrand]("")},
			{
				"wrong version",
				"550e8400-e29b-11d4-a716-446655440000",
				true,
				id.NewID[ComplaintBrand](""),
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				complaintID, err := ParseComplaintID(tt.input)

				if tt.wantErr {
					assert.Error(t, err)
					assert.True(t, complaintID.IsZero())
				} else {
					assert.NoError(t, err)
					assert.Equal(t, tt.expectedID, complaintID)
					assert.False(t, complaintID.IsZero())
					assert.Equal(t, tt.input, complaintID.String())
				}
			})
		}
	})
}
