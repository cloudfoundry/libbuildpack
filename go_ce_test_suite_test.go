package buildpack_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestGoCeTest(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GoCeTest Suite")
}
