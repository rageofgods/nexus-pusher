package comps

import (
	"strings"
)

// ComponentType Custom component type
type ComponentType string

func (c *ComponentType) Lower() ComponentType {
	return ComponentType(strings.ToLower(string(*c)))
}

const (
	// NPM Set NPM specific variables
	NPM    ComponentType = "npm"
	npmSrv string        = "https://registry.npmjs.org/"

	// PYPI Set PYPI specific variables
	PYPI    ComponentType = "pypi"
	pypiSrv string        = "https://pypi.org/"

	//NUGET  ComponentType = "nuget"
	//HELM   ComponentType = "helm"
	//DOCKER ComponentType = "docker"
	//RUBY   ComponentType = "rubygems"
	//APT    ComponentType = "apt"
)

type Typer interface {
	DownloadComponent() ([]byte, error)
	PrepareDataToUpload([]byte) (interface{}, error)
}
