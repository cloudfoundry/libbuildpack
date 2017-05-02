package libbuildpack_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	bp "github.com/cloudfoundry/libbuildpack"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Stager", func() {
	var (
		manifest    bp.Manifest
		buildDir    string
		cacheDir    string
		depsDir     string
		depsIdx     string
		logger      bp.Logger
		s           bp.Stager
		err         error
		oldCfStack  string
		buffer      *bytes.Buffer
		manifestDir string
	)

	BeforeEach(func() {
		buildDir, err = ioutil.TempDir("", "build")
		Expect(err).To(BeNil())

		cacheDir, err = ioutil.TempDir("", "cache")
		Expect(err).To(BeNil())

		depsDir, err = ioutil.TempDir("", "deps")
		Expect(err).To(BeNil())

		depsIdx = "0"
		err = os.MkdirAll(filepath.Join(depsDir, depsIdx), 0755)
		Expect(err).To(BeNil())

		manifestDir = filepath.Join("fixtures", "manifest", "standard")

		manifest, err = bp.NewManifest(manifestDir, time.Now())
		Expect(err).To(BeNil())

		logger = bp.NewLogger()
		logger.SetOutput(ioutil.Discard)
	})

	JustBeforeEach(func() {
		s = bp.Stager{
			BuildDir: buildDir,
			CacheDir: cacheDir,
			DepsDir:  depsDir,
			DepsIdx:  depsIdx,
			Manifest: manifest,
			Log:      logger,
		}
	})

	AfterEach(func() {
		err = os.RemoveAll(buildDir)
		Expect(err).To(BeNil())

		err = os.RemoveAll(cacheDir)
		Expect(err).To(BeNil())

		err = os.RemoveAll(depsDir)
		Expect(err).To(BeNil())
	})

	Describe("NewStager", func() {
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
			It("sets it in the Stager struct", func() {
				args = []string{"buildDir", "cacheDir", "depsDir", "idx"}
				s, err := bp.NewStager(args, bp.NewLogger())
				Expect(err).To(BeNil())
				Expect(s.BuildDir).To(Equal("buildDir"))
				Expect(s.CacheDir).To(Equal("cacheDir"))
				Expect(s.DepsDir).To(Equal("depsDir"))
				Expect(s.DepDir()).To(Equal("depsDir/idx"))
			})
		})

		Context("A deps dir is not provided", func() {
			It("sets DepsDir to the empty string", func() {
				args = []string{"buildDir", "cacheDir"}
				s, err := bp.NewStager(args, bp.NewLogger())
				Expect(err).To(BeNil())
				Expect(s.BuildDir).To(Equal("buildDir"))
				Expect(s.CacheDir).To(Equal("cacheDir"))
				Expect(s.DepsDir).To(Equal(""))
				Expect(s.DepDir()).To(Equal(""))
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

				_, err := bp.NewStager(args, logger)
				Expect(err).NotTo(BeNil())

				Expect(buffer.String()).To(ContainSubstring("Unable to load buildpack manifest"))
			})
		})
	})

	Describe("WriteConfigYml", func() {
		It("creates a file in the <depDir>/idx directory", func() {
			err := s.WriteConfigYml(nil)
			Expect(err).To(BeNil())

			contents, err := ioutil.ReadFile(filepath.Join(s.DepDir(), "config.yml"))
			Expect(err).To(BeNil())

			Expect(string(contents)).To(Equal("config: {}\nname: dotnet-core\n"))
		})

		It("writes passed config struct to file", func() {
			err := s.WriteConfigYml(map[string]string{"key":"value", "a":"b"})
			Expect(err).To(BeNil())

			contents, err := ioutil.ReadFile(filepath.Join(s.DepDir(), "config.yml"))
			Expect(err).To(BeNil())

			Expect(string(contents)).To(Equal("config:\n  a: b\n  key: value\nname: dotnet-core\n"))
		})
	})

	Describe("GetBuildpackDir", func() {
		var (
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
		BeforeEach(func() {
			oldCfStack = os.Getenv("CF_STACK")
			err = os.Setenv("CF_STACK", "cflinuxfs2")
			Expect(err).To(BeNil())
		})

		Context("buildpack is valid", func() {
			BeforeEach(func() {
				buffer = new(bytes.Buffer)
				logger = bp.NewLogger()
				logger.SetOutput(buffer)
			})

			It("it logs the buildpack name and version", func() {
				err := s.CheckBuildpackValid()
				Expect(err).To(BeNil())
				Expect(buffer.String()).To(ContainSubstring("-----> Dotnet-Core Buildpack version 99.99"))
			})
		})
	})

	Describe("ClearCache", func() {
		Context("already empty", func() {
			It("returns successfully", func() {
				err = s.ClearCache()
				Expect(err).To(BeNil())
				Expect(cacheDir).To(BeADirectory())
			})
		})

		Context("cache dir does not exist", func() {
			BeforeEach(func() {
				cacheDir = filepath.Join("not", "real")
			})

			It("returns successfully", func() {
				err = s.ClearCache()
				Expect(err).To(BeNil())
				Expect(cacheDir).ToNot(BeADirectory())
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
				err = s.ClearCache()
				Expect(err).To(BeNil())
				Expect(cacheDir).To(BeADirectory())

				fi, err := ioutil.ReadDir(cacheDir)
				Expect(err).To(BeNil())
				Expect(len(fi)).To(Equal(0))
			})
		})
	})

	Describe("ClearDepDir", func() {
		Context("already empty", func() {
			It("returns successfully", func() {
				err = s.ClearDepDir()
				Expect(err).To(BeNil())
				Expect(s.DepDir()).To(BeADirectory())
			})
		})

		Context("not empty", func() {
			BeforeEach(func() {
				Expect(os.MkdirAll(filepath.Join(depsDir, depsIdx, "fred", "jane"), 0755)).To(Succeed())
				Expect(ioutil.WriteFile(filepath.Join(depsDir, depsIdx, "fred", "jane", "jack.txt"), []byte("content"), 0644)).To(Succeed())
				Expect(ioutil.WriteFile(filepath.Join(depsDir, depsIdx, "jill.txt"), []byte("content"), 0644)).To(Succeed())
				Expect(ioutil.WriteFile(filepath.Join(depsDir, depsIdx, "config.yml"), []byte("yaml"), 0644)).To(Succeed())

				fi, err := ioutil.ReadDir(filepath.Join(depsDir, depsIdx))
				Expect(err).To(BeNil())
				Expect(len(fi)).To(Equal(3))
			})

			It("it clears the depDir, leaving config.yml", func() {
				err = s.ClearDepDir()
				Expect(err).To(BeNil())
				Expect(s.DepDir()).To(BeADirectory())

				fi, err := ioutil.ReadDir(s.DepDir())
				Expect(err).To(BeNil())
				Expect(len(fi)).To(Equal(1))

				content, err := ioutil.ReadFile(filepath.Join(s.DepDir(), "config.yml"))
				Expect(err).To(BeNil())
				Expect(string(content)).To(Equal("yaml"))
			})
		})
	})

	Describe("WriteEnvFile", func() {
		It("creates a file in the <depDir>/env directory", func() {
			err := s.WriteEnvFile("ENVVAR", "value")
			Expect(err).To(BeNil())

			contents, err := ioutil.ReadFile(filepath.Join(s.DepDir(), "env", "ENVVAR"))
			Expect(err).To(BeNil())

			Expect(string(contents)).To(Equal("value"))
		})
	})

	Describe("AddBinDependencyLink", func() {
		It("creates a symlink <depDir>/bin/<name> with the relative path to dest", func() {
			err := s.AddBinDependencyLink(filepath.Join(depsDir, depsIdx, "some", "long", "path"), "dep")
			Expect(err).To(BeNil())

			link, err := os.Readlink(filepath.Join(s.DepDir(), "bin", "dep"))
			Expect(err).To(BeNil())

			Expect(link).To(Equal("../some/long/path"))
		})
	})

	Describe("WriteProfileD", func() {
		var (
			info           os.FileInfo
			profileDScript string
			name           string
			contents       string
		)

		JustBeforeEach(func() {
			profileDScript = filepath.Join(s.DepDir(), "profile.d", name)

			err = s.WriteProfileD(name, contents)
			Expect(err).To(BeNil())
		})

		Context("profile.d directory exists", func() {
			BeforeEach(func() {
				name = "dir-exists.sh"
				contents = "used the dir"

				err = os.MkdirAll(filepath.Join(depsDir, depsIdx, "profile.d"), 0755)
				Expect(err).To(BeNil())
			})

			It("creates the file as an executable", func() {
				Expect(profileDScript).To(BeAnExistingFile())

				info, err = os.Stat(profileDScript)
				Expect(err).To(BeNil())

				// make sure at least 1 executable bit is set
				Expect(info.Mode().Perm() & 0111).NotTo(Equal(os.FileMode(0000)))
			})

			It("the script has the correct contents", func() {
				data, err := ioutil.ReadFile(profileDScript)
				Expect(err).To(BeNil())

				Expect(data).To(Equal([]byte("used the dir")))
			})
		})

		Context("profile.d directory does not exist", func() {
			BeforeEach(func() {
				name = "no-dir.sh"
				contents = "made the dir"
			})

			It("creates the file as an executable", func() {
				Expect(profileDScript).To(BeAnExistingFile())

				info, err = os.Stat(profileDScript)
				Expect(err).To(BeNil())

				// make sure at least 1 executable bit is set
				Expect(info.Mode().Perm() & 0111).NotTo(Equal(0000))
			})
		})

		It("the script has the correct contents", func() {
			data, err := ioutil.ReadFile(profileDScript)
			Expect(err).To(BeNil())

			Expect(data).To(Equal([]byte("made the dir")))
		})
	})
})
