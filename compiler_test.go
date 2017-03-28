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
	Describe("NewCompiler", func() {
		var (
			args     []string
			oldBpDir string
		)

		BeforeEach(func() {
			oldBpDir = os.Getenv("BUILDPACK_DIR")

			err := os.Setenv("BUILDPACK_DIR", "fixtures/manifest/standard")
			Expect(err).To(BeNil())
		})
		AfterEach(func() {
			err := os.Setenv("BUILDPACK_DIR", oldBpDir)
			Expect(err).To(BeNil())
		})

		Context("A deps dir is provided", func() {
			It("sets it in the compiler struct", func() {
				args = []string{"buildDir", "cacheDir", "", "depsDir"}
				c, err := bp.NewCompiler(args, bp.NewLogger())
				Expect(err).To(BeNil())
				Expect(c.BuildDir).To(Equal("buildDir"))
				Expect(c.CacheDir).To(Equal("cacheDir"))
				Expect(c.DepsDir).To(Equal("depsDir"))
			})
		})

		Context("A deps dir is not provided", func() {
			It("sets DepsDir to the empty string", func() {
				args = []string{"buildDir", "cacheDir"}
				c, err := bp.NewCompiler(args, bp.NewLogger())
				Expect(err).To(BeNil())
				Expect(c.BuildDir).To(Equal("buildDir"))
				Expect(c.CacheDir).To(Equal("cacheDir"))
				Expect(c.DepsDir).To(Equal(""))
			})
		})

		Context("the buildpack dir is invalid", func() {
			BeforeEach(func() {
				oldBpDir = os.Getenv("BUILDPACK_DIR")

				err := os.Setenv("BUILDPACK_DIR", "nothing/here")
				Expect(err).To(BeNil())
			})
			AfterEach(func() {
				err := os.Setenv("BUILDPACK_DIR", oldBpDir)
				Expect(err).To(BeNil())
			})

			It("returns an error and logs that it couldn't load the manifest", func() {
				args = []string{"buildDir", "cacheDir"}
				logger := bp.NewLogger()
				buffer := new(bytes.Buffer)
				logger.SetOutput(buffer)

				_, err := bp.NewCompiler(args, logger)
				Expect(err).NotTo(BeNil())

				Expect(buffer.String()).To(ContainSubstring("Unable to load buildpack manifest"))
			})
		})
	})

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

	Describe("ClearCache", func() {
		var (
			err      error
			cacheDir string
			compiler bp.Compiler
		)

		BeforeEach(func() {
			cacheDir, err = ioutil.TempDir("", "cache")
			Expect(err).To(BeNil())

		})

		JustBeforeEach(func() {
			compiler = bp.Compiler{BuildDir: "", CacheDir: cacheDir}
		})

		Context("already empty", func() {
			It("returns successfully", func() {
				err = compiler.ClearCache()
				Expect(err).To(BeNil())
				Expect(cacheDir).To(BeADirectory())
			})
		})

		Context("not empty", func() {
			BeforeEach(func() {
				Expect(os.MkdirAll(filepath.Join(cacheDir, "fred", "jane"), 0755)).To(Succeed())
				Expect(ioutil.WriteFile(filepath.Join(cacheDir, "fred", "jane", "jack.txt"), []byte("content"), 0644)).To(Succeed())
				Expect(ioutil.WriteFile(filepath.Join(cacheDir, "jill.txt"), []byte("content"), 0644)).To(Succeed())

				fi, err := ioutil.ReadDir(cacheDir)
				Expect(err).To(BeNil())
				Expect(len(fi)).To(Equal(2))
			})

			It("it clears the cache", func() {
				err = compiler.ClearCache()
				Expect(err).To(BeNil())
				Expect(cacheDir).To(BeADirectory())

				fi, err := ioutil.ReadDir(cacheDir)
				Expect(err).To(BeNil())
				Expect(len(fi)).To(Equal(0))
			})
		})
	})
})
