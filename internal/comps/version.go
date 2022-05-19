package comps

type Version struct {
	Version string
	Build   string
}

func NewVersion(version string, build string) *Version {
	return &Version{Version: version, Build: build}
}
