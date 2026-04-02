package domain

import (
	"regexp"

	"github.com/larsartmann/go-composable-business-types/id"
)

var agentIDValidation = newIDValidation(
	regexp.MustCompile(`^.{1,100}$`),
	"AgentID",
)

type AgentBrand struct{}

type AgentID = id.ID[AgentBrand, string]

func NewAgentID(name string) (AgentID, error) {
	return newBrandedID[AgentBrand](name, agentIDValidation)
}

func ParseAgentID(s string) (AgentID, error) {
	return parseBrandedID[AgentBrand](s, agentIDValidation)
}

func MustParseAgentID(s string) AgentID {
	return mustParseBrandedID[AgentBrand](s, agentIDValidation)
}
