package bdd_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Simple Test Setup", func() {
	It("should run a basic test", func() {
		Expect(true).To(BeTrue())
		Expect(1 + 1).To(Equal(2))
	})
})