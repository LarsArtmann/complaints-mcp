package domain

import (
	"regexp"

	"github.com/larsartmann/go-composable-business-types/id"
)

var projectIDValidation = newIDValidation(
	regexp.MustCompile(`^[a-zA-Z0-9\-_\s\.]{1,100}$`),
	"ProjectID",
)

type ProjectBrand struct{}

type ProjectID = id.ID[ProjectBrand, string]

func NewProjectID(name string) (ProjectID, error) {
	return newBrandedID[ProjectBrand](name, projectIDValidation)
}

func ParseProjectID(s string) (ProjectID, error) {
	return parseBrandedID[ProjectBrand](s, projectIDValidation)
}

func MustParseProjectID(s string) ProjectID {
	return mustParseBrandedID[ProjectBrand](s, projectIDValidation)
}
