package comps

import "strings"

// FileExtensionFromFile return file extension from uri path
func FileExtensionFromFile(fileName string) string {
	fileSplit := strings.Split(fileName, ".")
	return fileSplit[len(fileSplit)-1]
}

func pathSplit(assetPath string, index int) string {
	cmpPathSplit := strings.Split(assetPath, "/")
	return cmpPathSplit[len(cmpPathSplit)-index]
}

// AssetFileNameFromURI return file name from uri path
func AssetFileNameFromURI(assetPath string) string {
	return pathSplit(assetPath, 1)
}
