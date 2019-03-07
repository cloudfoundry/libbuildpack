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
	"gopkg.in/jarcoal/httpmock.v1"
)

var _ = Describe("Shims", func() {
	Describe("Installer", func() {
		Context("InstallCNBs", func() {
			BeforeEach(func() {
				Expect(os.Setenv("CF_STACK", "cflinuxfs3")).To(Succeed())

				httpmock.Reset()

				contents, err := ioutil.ReadFile(filepath.Join("fixtures", "buildpack", "bpA.tgz"))
				Expect(err).ToNot(HaveOccurred())

				httpmock.RegisterResponder("GET", "https://a-fake-url.com/bpA.tgz",
					httpmock.NewStringResponder(200, string(contents)))

				contents, err = ioutil.ReadFile(filepath.Join("fixtures", "buildpack", "bpB.tgz"))
				Expect(err).ToNot(HaveOccurred())

				httpmock.RegisterResponder("GET", "https://a-fake-url.com/bpB.tgz",
					httpmock.NewStringResponder(200, string(contents)))
			})

			AfterEach(func() {
				Expect(os.Unsetenv("CF_STACK")).To(Succeed())
			})

			It("installs the latest/unique buildpacks from an order.toml that are not already installed", func() {
				tmpDir, err := ioutil.TempDir("", "")
				Expect(err).ToNot(HaveOccurred())
				defer os.RemoveAll(tmpDir)
				Expect(os.MkdirAll(filepath.Join(tmpDir, "this.is.a.fake.bpC", "1.0.2"), 0777)).To(Succeed())

				buffer := new(bytes.Buffer)
				logger := libbuildpack.NewLogger(ansicleaner.New(buffer))

				manifest, err := libbuildpack.NewManifest(filepath.Join("fixtures", "buildpack"), logger, time.Now())
				Expect(err).To(BeNil())

				installer := shims.NewCNBInstaller(manifest)

				Expect(installer.InstallCNBS(filepath.Join("fixtures", "buildpack", "order.toml"), tmpDir)).To(Succeed())
				Expect(filepath.Join(tmpDir, "this.is.a.fake.bpA", "1.0.1", "a.txt")).To(BeAnExistingFile())
				Expect(filepath.Join(tmpDir, "this.is.a.fake.bpB", "1.0.2", "b.txt")).To(BeAnExistingFile())
				Expect(filepath.Join(tmpDir, "this.is.a.fake.bpA", "latest")).To(BeAnExistingFile())
				Expect(filepath.Join(tmpDir, "this.is.a.fake.bpB", "latest")).To(BeAnExistingFile())

				Expect(buffer.String()).ToNot(ContainSubstring("Installing this.is.a.fake.bpC"))
				Expect(filepath.Join(tmpDir, "this.is.a.fake.bpC")).To(BeADirectory())
			})
		})
	})
})
