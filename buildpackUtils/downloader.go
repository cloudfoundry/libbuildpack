package buildpackUtils

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
)

type Downloader struct {
	OutputDir string
	Manifest  *Manifest
}

func NewDownloader(dir string, manifest *Manifest) *Downloader {
	d := &Downloader{
		OutputDir: dir,
		Manifest:  manifest,
	}
	return d
}

func (d *Downloader) Fetch(dep Dependency) (string, error) {
	url, err := d.Manifest.GetUrl(dep)

	if err != nil {
		return "", err
	}

	resp, err := http.Get(url)
	defer resp.Body.Close()

	if err != nil {
		return "", err
	}

	blob, _ := ioutil.ReadAll(resp.Body)

	dest := path.Join(d.OutputDir, filenameFromUrl(url))

	err = os.Mkdir(d.OutputDir, os.ModePerm)

	if err != nil {
		return "", err
	}

	err = ioutil.WriteFile(dest, blob, os.ModePerm)

	if err != nil {
		return "", err
	}

	fmt.Printf("Downloaded [%s] to %s\n", url, dest)
	return dest, nil
}

func filenameFromUrl(url string) string {
	substrings := strings.Split(url, "/")
	return substrings[len(substrings)-1]
}

func (m *Manifest) GetUrl(dep Dependency) (string, error) {
	entry, err := m.GetEntry(dep)

	if err != nil {
		return "", err
	}

	return entry.URI, nil
}

func (m *Manifest) GetEntry(dep Dependency) (ManifestEntry, error) {
	var entry ManifestEntry
	inManifest := false

	for _, e := range m.ManifestEntries {
		if e.Dependency == dep {
			entry = e
			inManifest = true
			break
		}
	}

	if !inManifest {
		return entry, errors.New(fmt.Sprintf("dependency %s %s not found", dep.Name, dep.Version))
	}

	return entry, nil
}
