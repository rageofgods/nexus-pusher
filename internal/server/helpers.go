package server

import (
	"regexp"
)

// isValidNexusRepoName check for supported nexus repository name pattern
func isValidNexusRepoName(param string) bool {
	return regexp.MustCompile(`^[a-zA-Z\d_.-]+$`).MatchString(param)
}
