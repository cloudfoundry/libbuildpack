package packager_test

import (
	"archive/zip"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/cloudfoundry/libbuildpack/packager"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestPackager(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Packager Suite")
}

var _ = BeforeSuite(func() {
	packager.Stdout = GinkgoWriter
	packager.Stderr = GinkgoWriter
})

// Returns the list of files in a zip archive.
func ZipFiles(zipFile string) []string {
	r, err := zip.OpenReader(zipFile)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()

	filePaths := make([]string, 0, len(r.File))

	for _, f := range r.File {
		filePaths = append(filePaths, f.Name)
	}

	return filePaths
}

func ZipContents(zipFile, file string) (string, error) {
	r, err := zip.OpenReader(zipFile)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()

	for _, f := range r.File {
		if f.Name == file {
			rc, err := f.Open()
			if err != nil {
				return "", err
			}
			defer rc.Close()
			body, err := ioutil.ReadAll(rc)
			if err != nil {
				return "", err
			}
			return string(body), nil
		}
	}
	return "", fmt.Errorf("%s not found in %s", file, zipFile)
}
