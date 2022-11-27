package models

type Person struct {
	// The person's name.
	Name string `json:"name"`

	// The person's current certifications.
	Certifications []string `json:"certifications"`

	// The person's current status.
	Status AssetState `json:"status"`
}
