package config

import (
	"io"
	"net/http"
	"strings"
)

// ComponentType Custom component type
type ComponentType string

// Lower convert component type to lower registry
func (c ComponentType) Lower() ComponentType {
	return ComponentType(strings.ToLower(string(c)))
}

// String convert to string type
func (c ComponentType) String() string {
	return string(c)
}

// Bundled check if ComponentType must be processed as bundle of assets
// For example, maven2 requires to upload pom with assets simultaneously
func (c ComponentType) Bundled() bool {
	switch c.Lower() {
	case NPM:
		return false
	case PYPI:
		return false
	case MAVEN2:
		return true
	case NUGET:
		return false
	default:
		// Return false by default is safe here because we already
		// check component type in the previous code logic
		return false
	}
}

const (
	// NPM Set NPM specific variables
	NPM ComponentType = "npm"

	// PYPI Set PYPI specific variables
	PYPI ComponentType = "pypi"

	// MAVEN2 Set MAVEN2 specific variables
	MAVEN2 ComponentType = "maven2"

	// NUGET Set NUGET specific variables
	NUGET ComponentType = "nuget"

	// HELM   ComponentType = "helm"
	// DOCKER ComponentType = "docker"
	// RUBY   ComponentType = "rubygems"
	// APT    ComponentType = "apt"
)

type Asseter interface {
	DownloadAsset() (*http.Response, error)
	PrepareAssetToUpload(io.Reader) (string, io.Reader)
}

type Componenter interface {
	DownloadComponent() ([]*http.Response, error)
	PrepareComponentToUpload([]*http.Response) (string, io.Reader)
}
