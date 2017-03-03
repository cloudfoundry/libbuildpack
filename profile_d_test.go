package libbuildpack_test

import (
	"io/ioutil"
	"os"
	"path/filepath"

	bp "github.com/cloudfoundry/libbuildpack"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ProfileD", func() {
	var (
		info           os.FileInfo
		profileDScript string
		name           string
		contents       string
		buildDir       string
		err            error
	)
	BeforeEach(func() {
		buildDir, err = ioutil.TempDir("", "libbuildpack.test.builddir.")
		Expect(err).To(BeNil())
	})

	AfterEach(func() {
		err = os.RemoveAll(buildDir)
		Expect(err).To(BeNil())
	})

	JustBeforeEach(func() {
		profileDScript = filepath.Join(buildDir, ".profile.d", name)

		err = bp.WriteProfileD(buildDir, name, contents)
		Expect(err).To(BeNil())
	})

	Context(".profile.d directory exists", func() {
		BeforeEach(func() {
			name = "dir-exists.sh"
			contents = "used the dir"

			err = os.Mkdir(filepath.Join(buildDir, ".profile.d"), 0755)
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
	Context(".profile.d directory does not exist", func() {
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
