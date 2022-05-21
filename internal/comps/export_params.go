package comps

import "fmt"

type NexusExportComponents struct {
	NexusServer NexusServer             `json:"nexusServer"`
	Items       []*NexusExportComponent `json:"items"`
}

type NexusExportComponent struct {
	Name            string                       `json:"name"`
	Version         string                       `json:"version"`
	Repository      string                       `json:"repository"`
	Format          string                       `json:"format"`
	Group           string                       `json:"group"`
	ArtifactsSource string                       `json:"artifactsSource"`
	Assets          []*NexusExportComponentAsset `json:"assets"`
}

// FullName returns name and version for component
func (n NexusExportComponent) FullName() string {
	return fmt.Sprintf("%s-%s", n.Name, n.Version)
}

type NexusExportComponentAsset struct {
	Name        string `json:"name"`
	FileName    string `json:"fileName"`
	Version     string `json:"version"`
	Path        string `json:"path"`
	ContentType string `json:"contentType"`
}

// FullName returns name and version for asset
func (n NexusExportComponentAsset) FullName() string {
	return fmt.Sprintf("%s-%s", n.Name, n.Version)
}
