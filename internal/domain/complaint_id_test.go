package domain

import (
	"testing"

	brandedid "github.com/larsartmann/go-branded-id"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComplaintID_JSONSerialization_FlatStructure(t *testing.T) {
	t.Run("marshal produces flat JSON", func(t *testing.T) {
		AssertFlatJSONMarshaling(
			t,
			brandedid.NewID[ComplaintBrand]("550e8400-e29b-41d4-a716-446655440000"),
			"550e8400-e29b-41d4-a716-446655440000",
		)
	})

	t.Run("unmarshal from flat JSON", func(t *testing.T) {
		AssertUnmarshalFromFlatJSON(
			t,
			brandedid.NewID[ComplaintBrand]("550e8400-e29b-41d4-a716-446655440000"),
			`"550e8400-e29b-41d4-a716-446655440000"`,
		)
	})

	t.Run("reject nested JSON format", func(t *testing.T) {
		AssertRejectNestedJSONFormat[ComplaintBrand](
			t,
			`{"id":{"Value":"550e8400-e29b-41d4-a716-446655440000"}}`,
		)
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
			name        string
			input       string
			expectedID  ComplaintID
			wantErr     bool
		}{
			{
				name:        "valid UUID",
				input:       "550e8400-e29b-41d4-a716-446655440000",
				expectedID:  brandedid.NewID[ComplaintBrand]("550e8400-e29b-41d4-a716-446655440000"),
				wantErr:     false,
			},
			{
				name:        "valid lowercase",
				input:       "9cb3bb9e-b6dc-4e02-9767-e396a42b63a6",
				expectedID:  brandedid.NewID[ComplaintBrand]("9cb3bb9e-b6dc-4e02-9767-e396a42b63a6"),
				wantErr:     false,
			},
			{
				name:        "empty string",
				input:       "",
				expectedID:  brandedid.NewID[ComplaintBrand](""),
				wantErr:     true,
			},
			{
				name:        "invalid format",
				input:       "not-a-uuid",
				expectedID:  brandedid.NewID[ComplaintBrand](""),
				wantErr:     true,
			},
			{
				name:        "wrong version",
				input:       "550e8400-e29b-11d4-a716-446655440000",
				expectedID:  brandedid.NewID[ComplaintBrand](""),
				wantErr:     true,
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
