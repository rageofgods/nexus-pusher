package client

import (
	"nexus-pusher/internal/comps"
)

// genNexExpCompFromNexComp is converting original nexus structure data to compact export format
func genNexExpCompFromNexComp(c []*comps.NexusComponent) *comps.NexusExportComponents {
	var ec []*comps.NexusExportComponent
	for _, v := range c {
		var assets []*comps.NexusExportComponentAsset
		for _, vv := range v.Assets {
			exportAsset := &comps.NexusExportComponentAsset{Path: vv.Path, ContentType: vv.ContentType}
			assets = append(assets, exportAsset)
		}
		exportComponent := &comps.NexusExportComponent{
			Name:       v.Name,
			Version:    v.Version,
			Repository: v.Repository,
			Format:     v.Format,
			Assets:     assets}
		ec = append(ec, exportComponent)
	}
	return &comps.NexusExportComponents{Items: ec}
}
