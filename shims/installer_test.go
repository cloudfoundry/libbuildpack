package shims_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/cloudfoundry/libbuildpack"
	"github.com/cloudfoundry/libbuildpack/ansicleaner"
	"github.com/cloudfoundry/libbuildpack/shims"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	httpmock "gopkg.in/jarcoal/httpmock.v1"
)

var _ = Describe("Installer", func() {
	var (
		installer *shims.CNBInstaller
		tmpDir    string
		buffer    *bytes.Buffer
		err       error
	)

	BeforeEach(func() {
		Expect(os.Setenv("CF_STACK", "cflinuxfs3")).To(Succeed())
		httpmock.Reset()
		tmpDir, err = ioutil.TempDir("", "")
		Expect(err).ToNot(HaveOccurred())

		buffer = new(bytes.Buffer)
		logger := libbuildpack.NewLogger(ansicleaner.New(buffer))

		manifest, err := libbuildpack.NewManifest(filepath.Join("testdata", "buildpack"), logger, time.Now())
		Expect(err).To(BeNil())

		installer = shims.NewCNBInstaller(manifest)
	})

	AfterEach(func() {
		Expect(os.Unsetenv("CF_STACK")).To(Succeed())
		os.RemoveAll(tmpDir)
	})

	Context("InstallCNBs", func() {
		BeforeEach(func() {
			Expect(os.MkdirAll(filepath.Join(tmpDir, "this.is.a.fake.bpC", "1.0.2"), 0777)).To(Succeed())
			contents, err := ioutil.ReadFile(filepath.Join("testdata", "buildpack", "bpA.tgz"))
			Expect(err).ToNot(HaveOccurred())

			httpmock.RegisterResponder("GET", "https://a-fake-url.com/bpA.tgz",
				httpmock.NewStringResponder(200, string(contents)))

			contents, err = ioutil.ReadFile(filepath.Join("testdata", "buildpack", "bpB.tgz"))
			Expect(err).ToNot(HaveOccurred())

			httpmock.RegisterResponder("GET", "https://a-fake-url.com/bpB.tgz",
				httpmock.NewStringResponder(200, string(contents)))
		})

		It("installs the latest/unique buildpacks from an order.toml that are not already installed", func() {
			Expect(installer.InstallCNBS(filepath.Join("testdata", "buildpack", "order.toml"), tmpDir)).To(Succeed())
			Expect(filepath.Join(tmpDir, "this.is.a.fake.bpA", "1.0.1", "a.txt")).To(BeAnExistingFile())
			Expect(filepath.Join(tmpDir, "this.is.a.fake.bpB", "1.0.2", "b.txt")).To(BeAnExistingFile())
			Expect(filepath.Join(tmpDir, "this.is.a.fake.bpA", "latest")).To(BeAnExistingFile())
			Expect(filepath.Join(tmpDir, "this.is.a.fake.bpB", "latest")).To(BeAnExistingFile())

			Expect(buffer.String()).ToNot(ContainSubstring("Installing this.is.a.fake.bpC"))
			Expect(filepath.Join(tmpDir, "this.is.a.fake.bpC")).To(BeADirectory())
		})
	})

	Context("InstallLifecycle", func() {
		BeforeEach(func() {
			contents, err := ioutil.ReadFile(filepath.Join("testdata", "buildpack", "lifecycle-bundle.tgz"))
			Expect(err).ToNot(HaveOccurred())

			httpmock.RegisterResponder("GET", "https://a-fake-url.com/lifecycle-bundle.tgz",
				httpmock.NewStringResponder(200, string(contents)))
		})

		It("unpacks the lifecycle bundle and globs the contents of the subfolder and copies it somewhere", func() {
			Expect(installer.InstallLifecycle(tmpDir)).To(Succeed())
			keepBinaries := []string{"detector", "builder", "launcher"}

			for _, binary := range keepBinaries {
				Expect(filepath.Join(tmpDir, binary)).To(BeAnExistingFile())
			}
		})
	})
})
