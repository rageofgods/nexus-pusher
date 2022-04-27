package client

import (
	"github.com/goccy/go-json"
	"io/ioutil"
	"nexus-pusher/pkg/comps"
	"os"
	"strings"
)

func ReadExport(fileName string) (*comps.NexusExportComponents, error) {
	data := &comps.NexusExportComponents{}
	jsonFile, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}
	if err := jsonFile.Close(); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(byteValue, &data); err != nil {
		return nil, err
	}
	return data, nil
}

func componentNameFromPath(cmpPath string) string {
	cmpPathSplit := strings.Split(cmpPath, "/")
	return cmpPathSplit[len(cmpPathSplit)-1]
}
