package bdd_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestBDDSetup(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "BDD Test Suite", Label("bdd"))
}