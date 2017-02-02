package buildpack_test

import (
	"io/ioutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	be "github.com/sesmith177/buildpack-extensions"

	"testing"
)

func TestBuildpackExtensionsTest(t *testing.T) {
	be.Log.SetOutput(ioutil.Discard)

	RegisterFailHandler(Fail)
	RunSpecs(t, "BuildpackExtensionsTest Suite")
}
