package buildpackUtils

import (
	"errors"
	"fmt"
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
		return url, err
	}

	return url, nil
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
