package core

import (
	"strings"
	"time"
)

type NexusServer struct {
	Host             string
	BaseUrl          string
	ApiComponentsUrl string
	Username         string
	Password         string
}

func NewNexusServer(user string, pass string, host string, baseUrl string, apiComponentsUrl string) *NexusServer {
	return &NexusServer{
		Host:             host,
		BaseUrl:          baseUrl,
		ApiComponentsUrl: apiComponentsUrl,
		Username:         user,
		Password:         pass,
	}
}

//easyjson:json
type (
	NexusComponents struct {
		Items             []*NexusComponent `json:"items"`
		ContinuationToken string            `json:"continuationToken"`
	}

	NexusComponent struct {
		ID         string                 `json:"id"`
		Repository string                 `json:"repository"`
		Format     string                 `json:"format"`
		Group      string                 `json:"group"`
		Name       string                 `json:"name"`
		Version    string                 `json:"version"`
		Assets     []*NexusComponentAsset `json:"assets"`
	}

	NexusComponentAsset struct {
		DownloadURL string `json:"downloadUrl"`
		Path        string `json:"path"`
		ID          string `json:"id"`
		Repository  string `json:"repository"`
		Format      string `json:"format"`
		Checksum    struct {
			AdditionalProp1 struct {
			} `json:"additionalProp1"`
			AdditionalProp2 struct {
			} `json:"additionalProp2"`
			AdditionalProp3 struct {
			} `json:"additionalProp3"`
		} `json:"checksum"`
		ContentType  string    `json:"contentType"`
		LastModified time.Time `json:"lastModified"`
	}
)

// AssetPathWithoutTrailingZeroes will remove trailing zeroes from component
// path. It's used to workaround nexus behavior with multiple assets
// with same version but different trailing zero count.
func (nca NexusComponentAsset) AssetPathWithoutTrailingZeroes() string {
	resultPath := nca.Path
	suffix := ".0"
	for strings.HasSuffix(resultPath, suffix) {
		resultPath = strings.TrimSuffix(resultPath, suffix)
	}
	return resultPath
}

type UploadResult struct {
	ComponentPath string
	Err           error
}
