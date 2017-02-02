package buildpack_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestBuildpackExtensionsTest(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "BuildpackExtensionsTest Suite")
}
