package domain

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/larsartmann/go-branded-id"
)

const maxIDLength = 100

type idValidation struct {
	pattern  *regexp.Regexp
	typeName string
}

func newIDValidation(pattern *regexp.Regexp, typeName string) idValidation {
	return idValidation{pattern: pattern, typeName: typeName}
}

func (v idValidation) validate(s string) error {
	if s == "" {
		return errors.New("cannot be empty")
	}

	if len(s) > maxIDLength {
		return errors.New("cannot exceed 100 characters")
	}

	if !v.pattern.MatchString(s) {
		return errors.New("contains invalid characters")
	}

	return nil
}

func newBrandedID[Brand any](name string, validation idValidation) (go-branded-id.ID[Brand, string], error) {
	trimmed := strings.TrimSpace(name)

	err := validation.validate(trimmed)
	if err != nil {
		return go-branded-id.NewID[Brand](""), fmt.Errorf("invalid %s: %w", validation.typeName, err)
	}

	return go-branded-id.NewID[Brand](trimmed), nil
}

func parseBrandedID[Brand any](s string, validation idValidation) (go-branded-id.ID[Brand, string], error) {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return go-branded-id.NewID[Brand](""), nil
	}

	err := validation.validate(trimmed)
	if err != nil {
		return go-branded-id.NewID[Brand](""), fmt.Errorf("invalid %s: %w", validation.typeName, err)
	}

	return go-branded-id.NewID[Brand](trimmed), nil
}

func mustParseBrandedID[Brand any](s string, validation idValidation) go-branded-id.ID[Brand, string] {
	brandedID, err := parseBrandedID[Brand](s, validation)
	if err != nil {
		panic(fmt.Sprintf("invalid %s: %s", validation.typeName, err))
	}

	return brandedID
}

func ValidateAgentID(s string) error {
	return agentIDValidation.validate(s)
}

func ValidateSessionID(s string) error {
	return sessionIDValidation.validate(s)
}

func ValidateProjectID(s string) error {
	return projectIDValidation.validate(s)
}

func ValidateOptionalID[Brand any](
	brandedID go-branded-id.ID[Brand, string],
	typeName string,
	validateFn func(string) error,
) error {
	if !brandedID.IsZero() {
		err := validateFn(brandedID.Get())
		if err != nil {
			return fmt.Errorf("invalid %s: %w", typeName, err)
		}
	}

	return nil
}
