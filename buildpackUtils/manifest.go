package buildpackUtils

type Dependency struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}

type ManifestEntry struct {
	Dependency Dependency `yaml:",inline"`
	URI        string     `yaml:"uri"`
	MD5        string     `yaml:"md5"`
	CFStacks   []string   `yam:"cf_stacks"`
}

type URLToDependencyMap struct {
	Match   string `yaml:"match"`
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}

type Manifest struct {
	Language        string               `yaml:"language"`
	DefaultVersions []Dependency         `yaml:"default_versions"`
	URLMaps         []URLToDependencyMap `yaml:"url_to_dependency_map"`
	ManifestEntries []ManifestEntry      `yaml:"dependencies"`
	ExcludeFiles    []string             `yaml:"exclude_files"`
}
