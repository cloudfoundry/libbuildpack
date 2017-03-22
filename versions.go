package libbuildpack

import (
	"fmt"
	"sort"

	"github.com/Masterminds/semver"
)

func FindMatchingVersion(versionSpec string, existingVersions []string) (string, error) {
	constraint, err := semver.NewConstraint(versionSpec)
	if err != nil {
		return "", err
	}

	var matching semver.Collection

	for _, ver := range existingVersions {
		v, err := semver.NewVersion(ver)
		if err != nil {
			return "", err
		}
		if constraint.Check(v) {
			matching = append(matching, v)
		}
	}

	if len(matching) > 0 {
		sort.Sort(matching)
		max := matching[len(matching)-1].String()
		return max, nil
	}

	return "", fmt.Errorf("no match found for %s in %v", versionSpec, existingVersions)
}
