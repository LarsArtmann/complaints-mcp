package domain

import (
	"regexp"

	"github.com/larsartmann/go-branded-id"
)

type (
	AgentBrand   struct{}
	ProjectBrand struct{}
	SessionBrand struct{}
)

type (
	AgentID   = id.ID[AgentBrand, string]
	ProjectID = id.ID[ProjectBrand, string]
	SessionID = id.ID[SessionBrand, string]
)

type ComplaintIDField int

const (
	ComplaintFieldAgentID ComplaintIDField = iota
	ComplaintFieldProjectID
	ComplaintFieldSessionID
)

// GetID extracts the string ID value based on the specified field type.
func (c *Complaint) GetID(field ComplaintIDField) string {
	switch field {
	case ComplaintFieldAgentID:
		return c.AgentID.String()
	case ComplaintFieldProjectID:
		return c.ProjectID.String()
	case ComplaintFieldSessionID:
		return c.SessionID.String()
	default:
		return ""
	}
}

var (
	agentIDValidation   = newIDValidation(regexp.MustCompile(`^.{1,100}$`), "AgentID")
	projectIDValidation = newIDValidation(
		regexp.MustCompile(`^[a-zA-Z0-9\-_\s\.]{1,100}$`),
		"ProjectID",
	)
	sessionIDValidation = newIDValidation(
		regexp.MustCompile(`^[a-zA-Z0-9\-_\s]{1,100}$`),
		"SessionID",
	)
)

func NewAgentID(name string) (AgentID, error) {
	return newBrandedID[AgentBrand](name, agentIDValidation)
}

func ParseAgentID(s string) (AgentID, error) {
	return parseBrandedID[AgentBrand](s, agentIDValidation)
}

func MustParseAgentID(s string) AgentID {
	return mustParseBrandedID[AgentBrand](s, agentIDValidation)
}

func NewProjectID(name string) (ProjectID, error) {
	return newBrandedID[ProjectBrand](name, projectIDValidation)
}

func ParseProjectID(s string) (ProjectID, error) {
	return parseBrandedID[ProjectBrand](s, projectIDValidation)
}

func MustParseProjectID(s string) ProjectID {
	return mustParseBrandedID[ProjectBrand](s, projectIDValidation)
}

func NewSessionID(name string) (SessionID, error) {
	return newBrandedID[SessionBrand](name, sessionIDValidation)
}

func ParseSessionID(s string) (SessionID, error) {
	return parseBrandedID[SessionBrand](s, sessionIDValidation)
}

func MustParseSessionID(s string) SessionID {
	return mustParseBrandedID[SessionBrand](s, sessionIDValidation)
}
