package client

import "nexus-pusher/pkg/comps"

//func ExportComponents(c []*comps.NexusComponent) error {
//	if err := writeExport(genNexExpCompFromNexComp(c), "export.json"); err != nil {
//		return err
//	}
//	return nil
//}
//
//func writeExport(ec *comps.NexusExportComponents, fileName string) error {
//	fmt.Printf("\nStarting export data to '%s'", fileName)
//	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
//	if err != nil {
//		return err
//	}
//	encoder := json.NewEncoder(file)
//	encoder.SetIndent("", " ")
//	if err := encoder.Encode(ec); err != nil {
//		return err
//	}
//	if err := file.Close(); err != nil {
//		return err
//	}
//	fmt.Printf("\nSuccesfully export data to '%s'", fileName)
//	return nil
//}
//

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
