package comps

type NexusExportComponents struct {
	NexusServer NexusServer             `json:"nexusServer"`
	Items       []*NexusExportComponent `json:"items"`
}

type NexusExportComponent struct {
	Name       string                       `json:"name"`
	Version    string                       `json:"version"`
	Repository string                       `json:"repository"`
	Format     string                       `json:"format"`
	Assets     []*NexusExportComponentAsset `json:"assets"`
}

type NexusExportComponentAsset struct {
	Name        string `json:"name"`
	FileName    string `json:"fileName"`
	Version     string `json:"version"`
	Path        string `json:"path"`
	ContentType string `json:"contentType"`
}
