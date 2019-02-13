package shims_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/libbuildpack/shims"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Releaser", func() {
	var (
		releaser shims.Releaser
		v2AppDir string
		buf      *bytes.Buffer
	)

	BeforeEach(func() {
		var err error

		v2AppDir, err = ioutil.TempDir("", "build")
		Expect(err).NotTo(HaveOccurred())

		contents := `
				buildpacks = ["some.buildpacks", "some.other.buildpack"]
				[[processes]]
				type = "web"
				command = "npm start"
				`
		Expect(os.MkdirAll(filepath.Join(v2AppDir, ".cloudfoundry"), 0777)).To(Succeed())
		Expect(ioutil.WriteFile(filepath.Join(v2AppDir, ".cloudfoundry", "metadata.toml"), []byte(contents), 0666)).To(Succeed())

		buf = &bytes.Buffer{}

		releaser = shims.Releaser{
			MetadataPath: filepath.Join(v2AppDir, ".cloudfoundry", "metadata.toml"),
			Writer:       buf,
		}
	})

	AfterEach(func() {
		Expect(os.RemoveAll(v2AppDir)).To(Succeed())
	})

	It("runs with the correct arguments and moves things to the correct place", func() {
		Expect(releaser.Release()).To(Succeed())
		Expect(buf.Bytes()).To(Equal([]byte("default_process_types:\n  web: npm start\n")))
		Expect(filepath.Join(v2AppDir, ".cloudfoundry", "metadata.toml")).NotTo(BeAnExistingFile())
	})
})
