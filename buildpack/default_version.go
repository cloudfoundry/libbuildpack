package buildpack

import (
	"fmt"
)

func (m *Manifest) DefaultVersionFor(depName string) (string, error) {
	var defaultVersion string
	numDefaults := 0

	for _, dep := range m.DefaultVersions {
		if depName == dep.Name {
			defaultVersion = dep.Version
			numDefaults++
		}
	}

	if numDefaults == 0 {
		return "", fmt.Errorf("no default version for: %s", depName)
	} else if numDefaults > 1 {
		return "", fmt.Errorf("found %d default versions for %s", numDefaults, depName)
	}

	return defaultVersion, nil

}

const DefaultVersionsError = "The buildpack manifest is misconfigured for 'default_versions'. " +
	"Contact your Cloud Foundry operator/admin. For more information, see " +
	"https://docs.cloudfoundry.org/buildpacks/custom.html#specifying-default-versions"
