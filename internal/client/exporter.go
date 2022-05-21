package client

import (
	"nexus-pusher/internal/comps"
)

// genNexExpCompFromNexComp is converting original nexus structure data to compact export format
func genNexExpCompFromNexComp(artifactsSource string, c []*comps.NexusComponent) *comps.NexusExportComponents {
	ec := make([]*comps.NexusExportComponent, 0, len(c))
	for _, v := range c {
		var assets []*comps.NexusExportComponentAsset
		for _, vv := range v.Assets {
			exportAsset := &comps.NexusExportComponentAsset{
				Name:        v.Name,
				Version:     v.Version,
				FileName:    func() string { return comps.AssetFileNameFromURI(vv.Path) }(),
				Path:        vv.Path,
				ContentType: vv.ContentType}
			assets = append(assets, exportAsset)
		}
		exportComponent := &comps.NexusExportComponent{
			Name:            v.Name,
			Version:         v.Version,
			Repository:      v.Repository,
			Format:          v.Format,
			Group:           v.Group,
			ArtifactsSource: artifactsSource,
			Assets:          assets,
		}
		ec = append(ec, exportComponent)
	}
	return &comps.NexusExportComponents{Items: ec}
}
