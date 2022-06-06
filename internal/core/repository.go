package core

// NexusRepository struct to hold nexus repo
type NexusRepository struct {
	Name       string              `json:"name"`
	Format     string              `json:"format"`
	Type       string              `json:"type"`
	URL        string              `json:"url"`
	Attributes NexusRepoAttributes `json:"attributes"`
}

// NexusRepoAttributes holds repo attributes struct
type NexusRepoAttributes struct {
	AdditionalProp1 struct {
	} `json:"additionalProp1"`
	AdditionalProp2 struct {
	} `json:"additionalProp2"`
	AdditionalProp3 struct {
	} `json:"additionalProp3"`
}
