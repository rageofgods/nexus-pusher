package core

import (
	"fmt"
	"io"
	"net/http"
	"nexus-pusher/pkg/http_clients"
	"nexus-pusher/pkg/utils"
	"strings"
)

type Maven2 struct {
	Server    string
	Component *NexusExportComponent
}

func NewMaven2(server string, component *NexusExportComponent) *Maven2 {
	return &Maven2{
		Server:    server,
		Component: component,
	}
}

// filterExtensions will remove hash assets from upload list
// because nexus server will create them dynamically upon uploading
func (m *Maven2) filterExtensions() {
	// Define extensions to filter
	mavenFilteredExtensions := map[string]struct{}{
		"sha1":   {},
		"md5":    {},
		"sha256": {},
		"sha512": {},
	}

	// Allocate memory for assets (max - current assets count)
	assets := make([]*NexusExportComponentAsset, 0, len(m.Component.Assets))

	for _, asset := range m.Component.Assets {
		fileExtension := FileExtensionFromFile(asset.FileName)
		// If asset file extension not in exceptions list add it to assets slice
		if _, ok := mavenFilteredExtensions[fileExtension]; !ok {
			assets = append(assets, asset)
		}
	}
	// Overwrite initial assets slice with filtered one
	m.Component.Assets = assets
}

func (m Maven2) DownloadComponent() ([]*http.Response, error) {
	// Allocate slice for responses following assets count
	responses := make([]*http.Response, 0, len(m.Component.Assets))

	for i := range m.Component.Assets {
		req, err := http.NewRequest("GET", m.assetDownloadURL(i), nil)
		if err != nil {
			return nil, fmt.Errorf("DownloadComponent: %w", err)
		}
		req.Header.Set("Accept", "application/octet-stream")

		// Send request
		resp, err := http_clients.HttpRetryClient(180).Do(req)
		if err != nil {
			return nil, fmt.Errorf("DownloadComponent: %w", err)
		}

		responses = append(responses, resp)
	}
	// Return all responses
	return responses, nil
}

func (m *Maven2) PrepareComponentToUpload(responses []*http.Response) (string, io.Reader) {
	// Create random boundary id
	boundary := utils.GenRandomBoundary(32)
	// Create slice of reader with length of responses
	// multiplied by 2 (format + binary data) plus
	// one element - bodyBottom closer
	readers := make([]io.Reader, 0, len(responses)*2+2)

	const fileHeader = "Content-type: application/octet-stream"
	const fileFormat = "--%s\r\nContent-Disposition: form-data; name=\"%s\"; filename=\"%s\"\r\n%s\r\n\r\n"
	const fieldFormat = "--%s\r\nContent-Disposition: form-data; name=\"%s\"\r\n\r\n%s\r\n"
	const fileTypeFormat = "maven2.asset%d"
	const fileExtensionFormat = "%s.extension"
	const fileClassifierFormat = "%s.classifier"
	const fieldGeneratePom = "maven2.generate-pom"
	const fieldArtifactID = "maven2.artifactId"
	const fieldGroupIdID = "maven2.groupId"
	const fieldVersion = "maven2.version"

	// We don't want to allow nexus to generate pom
	// to match original repository assets structure
	pomGenerationPart := fmt.Sprintf(fieldFormat, boundary, fieldGeneratePom, "false")
	readers = append(readers, strings.NewReader(pomGenerationPart))

	// Add required fields if we don't have pom in assets list
	if !m.pomInComponent() {
		artifactIdPart := fmt.Sprintf(fieldFormat, boundary, fieldArtifactID, m.Component.Name)
		groupIdPart := fmt.Sprintf(fieldFormat, boundary, fieldGroupIdID, m.Component.Group)
		versionPart := fmt.Sprintf(fieldFormat, boundary, fieldVersion, m.Component.Version)

		readers = append(readers,
			strings.NewReader(artifactIdPart),
			strings.NewReader(groupIdPart),
			strings.NewReader(versionPart))
	}

	// Iterate over all assets and form resulting body
	for i, resp := range responses {
		// Setup maven2 asset index (starting from 1) field
		fileTypeFormat := fmt.Sprintf(fileTypeFormat, i+1)
		// Setup maven2 asset extension field
		fileExtensionFormat := fmt.Sprintf(fileExtensionFormat, fileTypeFormat)
		// Setup maven2 asset classifier field
		fileClassifierFormat := fmt.Sprintf(fileClassifierFormat, fileTypeFormat)

		fileName := AssetFileNameFromURI(resp.Request.URL.Path)
		fileExtension := FileExtensionFromFile(fileName)
		fileClassifier := m.assetClassifier(fileName, fileExtension)

		// Generate extension part
		extensionPart := fmt.Sprintf(fieldFormat, boundary, fileExtensionFormat, fileExtension)
		// Generate file part
		filePart := fmt.Sprintf(fileFormat, boundary, fileTypeFormat, fileName, fileHeader)
		// Generate classifier part if exists
		classifierPart := ""
		if fileClassifier != "" {
			classifierPart = fmt.Sprintf(fieldFormat, boundary, fileClassifierFormat, fileClassifier)
		}
		// Combine all parts in one bundle
		combinedParts := fmt.Sprintf("%s%s%s", extensionPart, classifierPart, filePart)
		readers = append(readers, strings.NewReader(combinedParts), resp.Body, strings.NewReader("\r\n"))
	}

	// Close boundary
	bodyBottom := fmt.Sprintf("--%s--\r\n", boundary)
	readers = append(readers, strings.NewReader(bodyBottom))

	// Form body for http request
	body := io.MultiReader(readers...)
	contentType := fmt.Sprintf("multipart/form-data; boundary=%s", boundary)
	return contentType, body
}

func (m Maven2) assetDownloadURL(index int) string {
	return fmt.Sprintf("%s%s", m.Server, m.Component.Assets[index].Path)
}

func (m Maven2) assetClassifier(fileName string, fileExtension string) string {
	leftTrim := strings.TrimPrefix(fileName, fmt.Sprintf("%s-%s-", m.Component.Name, m.Component.Version))
	if leftTrim == fileName {
		return ""
	}
	result := strings.TrimSuffix(leftTrim, fmt.Sprintf(".%s", fileExtension))

	return result
}

// pomInComponent check if pom file is exists in assets list for current component
func (m Maven2) pomInComponent() bool {
	for _, asset := range m.Component.Assets {
		if FileExtensionFromFile(asset.FileName) == "pom" {
			return true
		}
	}
	return false
}
