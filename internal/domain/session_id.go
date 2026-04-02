package domain

import (
	"regexp"

	"github.com/larsartmann/go-composable-business-types/id"
)

var sessionIDValidation = newIDValidation(
	regexp.MustCompile(`^[a-zA-Z0-9\-_\s]{1,100}$`),
	"SessionID",
)

type SessionBrand struct{}

type SessionID = id.ID[SessionBrand, string]

func NewSessionID(name string) (SessionID, error) {
	return newBrandedID[SessionBrand](name, sessionIDValidation)
}

func ParseSessionID(s string) (SessionID, error) {
	return parseBrandedID[SessionBrand](s, sessionIDValidation)
}

func MustParseSessionID(s string) SessionID {
	return mustParseBrandedID[SessionBrand](s, sessionIDValidation)
}
