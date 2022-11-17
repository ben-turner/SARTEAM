package sartopo

type ShapeProperties struct {
	Title         string  `json:"title"`
	Description   string  `json:"description"`
	FolderID      string  `json:"folderId"`
	GPSType       string  `json:"gpstype"`
	StrokeWidth   int     `json:"stroke-width"`
	StrokeOpacity float64 `json:"stroke-opacity"`
	Stroke        string  `json:"stroke"`
	Pattern       string  `json:"pattern"`
	Fill          string  `json:"fill"`
	Class         string  `json:"class,omitempty"`
}

type ShapeGeometry struct {
	Type        string      `json:"type"`
	Coordinates [][]float64 `json:"coordinates"`
}

type Shape struct {
	Type       string          `json:"type,omitempty"`
	ID         string          `json:"id,omitempty"`
	Properties ShapeProperties `json:"properties"`
	Geometry   ShapeGeometry   `json:"geometry"`
}
