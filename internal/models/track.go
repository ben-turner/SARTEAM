package models

import "sort"

type Track struct {
	// The ID of the track.
	ID string `json:"id"`

	// The ID of the track in SARTopo.
	SARTopoID string `json:"sartopoId"`

	// The name of the track.
	Name string `json:"name"`

	// The description of the track.
	Description string `json:"description"`

	// The track points.
	Points PointList `json:"points"`
}

func (t *Track) AddPoint(p *Point) {
	t.Points = append(t.Points, p)
	sort.Sort(t.Points)
}

func (t *Track) PointsAs2DArray() [][]float64 {
	points := make([][]float64, len(t.Points))
	for i, p := range t.Points {
		points[i] = []float64{p.Longitude, p.Latitude}
	}

	return points
}
