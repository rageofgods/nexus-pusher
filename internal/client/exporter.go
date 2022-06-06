package client

import (
	"nexus-pusher/internal/core"
)

// genNexExpCompFromNexComp is converting original nexus structure data to compact export format
func genNexExpCompFromNexComp(artifactsSource string, c []*core.NexusComponent) *core.NexusExportComponents {
	ec := make([]*core.NexusExportComponent, 0, len(c))
	for _, v := range c {
		var assets []*core.NexusExportComponentAsset
		for _, vv := range v.Assets {
			exportAsset := &core.NexusExportComponentAsset{
				Name:        v.Name,
				Version:     v.Version,
				FileName:    func() string { return core.AssetFileNameFromURI(vv.Path) }(),
				Path:        vv.Path,
				ContentType: vv.ContentType}
			assets = append(assets, exportAsset)
		}
		exportComponent := &core.NexusExportComponent{
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
	return &core.NexusExportComponents{Items: ec}
}
