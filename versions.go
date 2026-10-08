package libbuildpack

import (
	"fmt"
	"sort"
	"strings"

	semver2 "github.com/Masterminds/semver"
	semver1 "github.com/blang/semver"
)

type versionWithOriginal struct {
	original string
	version  semver1.Version
}
type versionsWithOriginal []versionWithOriginal

func (v versionsWithOriginal) Len() int           { return len(v) }
func (v versionsWithOriginal) Swap(i, j int)      { v[i], v[j] = v[j], v[i] }
func (v versionsWithOriginal) Less(i, j int) bool { return v[i].version.LT(v[j].version) }

// normalizeSemver truncates a 4-part version string (e.g. "21.0.12.1") to its
// first three segments ("21.0.12") so that standard semver libraries can parse
// it. The original string is preserved separately for output.
func normalizeSemver(ver string) string {
	parts := strings.SplitN(ver, ".", 5)
	if len(parts) > 3 {
		return strings.Join(parts[:3], ".")
	}
	return ver
}

func FindMatchingVersion(constraint string, versions []string) (string, error) {
	vs, err := FindMatchingVersions(constraint, versions)
	if err != nil {
		return "", err
	}
	return vs[len(vs)-1], nil
}

func FindMatchingVersions(constraint string, versions []string) ([]string, error) {
	matchedVersions, err := matchSemver1(constraint, versions)
	if err == nil {
		return matchedVersions, nil
	}

	return matchSemver2(constraint, versions)
}

func matchSemver1(constraint string, versions []string) ([]string, error) {
	var depVersions versionsWithOriginal
	versionConstraint, err := semver1.ParseRange(normalizeSemver(constraint))
	if err != nil {
		return []string{}, err
	}

	for _, ver := range versions {
		depVersion, err := semver1.Parse(normalizeSemver(ver))
		if err != nil {
			return []string{}, err
		}
		versionWithOriginal := versionWithOriginal{
			original: ver,
			version:  depVersion,
		}

		if versionConstraint(depVersion) {
			depVersions = append(depVersions, versionWithOriginal)
		}
	}

	if len(depVersions) != 0 {
		sort.Sort(depVersions)
		var vs []string
		for _, depV := range depVersions {
			vs = append(vs, depV.original)
		}
		return vs, nil
	}

	return []string{}, fmt.Errorf("no match found for %s in %v", constraint, versions)
}

func matchSemver2(constraint string, versions []string) ([]string, error) {
	type versionEntry struct {
		original string
		parsed   *semver2.Version
	}
	var depVersions []versionEntry
	versionConstraint, err := semver2.NewConstraint(normalizeSemver(constraint))
	if err != nil {
		return []string{}, err
	}

	for _, ver := range versions {
		depVersion, err := semver2.NewVersion(normalizeSemver(ver))
		if err != nil {
			return []string{}, err
		}

		if versionConstraint.Check(depVersion) {
			depVersions = append(depVersions, versionEntry{original: ver, parsed: depVersion})
		}
	}

	if len(depVersions) != 0 {
		sort.Slice(depVersions, func(i, j int) bool {
			return depVersions[i].parsed.LessThan(depVersions[j].parsed)
		})
		var vs []string
		for _, e := range depVersions {
			vs = append(vs, e.original)
		}
		return vs, nil
	}

	return []string{}, fmt.Errorf("no match found for %s in %v", constraint, versions)
}
