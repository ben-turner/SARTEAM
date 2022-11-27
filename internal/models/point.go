package models

import "time"

type Point struct {
	// The latitude of the point.
	Latitude float64 `json:"latitude"`
	// The longitude of the point.
	Longitude float64 `json:"longitude"`
	// The altitude of the point.
	Altitude float64 `json:"altitude"`
	// The time the point was recorded.
	Time time.Time `json:"time"`
}

type PointList []*Point

func (p PointList) Len() int {
	return len(p)
}

func (p PointList) Less(i, j int) bool {
	return p[i].Time.Before(p[j].Time)
}

func (p PointList) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}
