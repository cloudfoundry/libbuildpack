package libbuildpack_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"

	bp "github.com/cloudfoundry/libbuildpack"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Compiler", func() {
	Describe("GetBuildpackDir", func() {
		var (
			err       error
			parentDir string
			testBpDir string
			oldBpDir  string
		)
		BeforeEach(func() {
			parentDir, err = filepath.Abs(filepath.Join(filepath.Dir(os.Args[0]), ".."))
			Expect(err).To(BeNil())
		})

		JustBeforeEach(func() {
			oldBpDir = os.Getenv("BUILDPACK_DIR")
			err = os.Setenv("BUILDPACK_DIR", testBpDir)
			Expect(err).To(BeNil())

		})

		AfterEach(func() {
			err = os.Setenv("BUILDPACK_DIR", oldBpDir)
			Expect(err).To(BeNil())
		})

		Context("BUILDPACK_DIR is set", func() {
			BeforeEach(func() {
				testBpDir = "buildpack_root_directory"
			})
			It("returns the value for BUILDPACK_DIR", func() {
				dir, err := bp.GetBuildpackDir()
				Expect(err).To(BeNil())
				Expect(dir).To(Equal("buildpack_root_directory"))
			})
		})
		Context("BUILDPACK_DIR is not set", func() {
			BeforeEach(func() {
				testBpDir = ""
			})
			It("returns the parent of the directory containing the executable", func() {
				dir, err := bp.GetBuildpackDir()
				Expect(err).To(BeNil())
				Expect(dir).To(Equal(parentDir))
			})
		})
	})

	Describe("CheckBuildpackValid", func() {
		var (
			manifest   bp.Manifest
			cacheDir   string
			logger     bp.Logger
			compiler   bp.Compiler
			err        error
			oldCfStack string
			buffer     *bytes.Buffer
		)

		BeforeEach(func() {
			oldCfStack = os.Getenv("CF_STACK")
			err = os.Setenv("CF_STACK", "cflinuxfs2")
			Expect(err).To(BeNil())

			cacheDir, err = ioutil.TempDir("", "cache")
			Expect(err).To(BeNil())

			manifest, err = bp.NewManifest("fixtures/manifest/standard")
			Expect(err).To(BeNil())

			logger = bp.NewLogger()
			logger.SetOutput(ioutil.Discard)
		})

		JustBeforeEach(func() {
			compiler = bp.Compiler{BuildDir: "", CacheDir: cacheDir, Manifest: manifest, Log: logger}
		})

		Context("buildpack is valid", func() {
			BeforeEach(func() {
				buffer = new(bytes.Buffer)
				logger = bp.NewLogger()
				logger.SetOutput(buffer)
			})

			It("it logs the buildpack name and version", func() {
				err := compiler.CheckBuildpackValid()
				Expect(err).To(BeNil())
				Expect(buffer.String()).To(ContainSubstring("-----> Dotnet-Core Buildpack version 99.99"))
			})
		})
	})
})
