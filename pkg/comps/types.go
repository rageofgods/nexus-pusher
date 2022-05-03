package comps

import "strings"

// ComponentType Custom component type
type ComponentType string

func (c *ComponentType) Lower() ComponentType {
	return ComponentType(strings.ToLower(string(*c)))
}

const (
	// NPM Set NPM specific variables
	NPM    ComponentType = "npm"
	npmSrv string        = "https://registry.npmjs.org/"
)
