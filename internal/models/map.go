package models

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/ben-turner/sarteam/internal/sartopo"
)

type Map struct {
	// The address at which sartopo is running.
	SARTopoAddress string `json:"address"`

	// The ID of the map.
	ID string `json:"id"`

	// A mapping of sarteam track IDs to sartopo track IDs.
	TrackIDs map[string]string `json:"trackIds"`
}

func (m *Map) post(path string, body string) error {
	endpoint := m.SARTopoAddress + "/api/v1/map/" + m.ID + path

	resp, err := http.PostForm(endpoint, url.Values{
		"json": {body}, // Why send json when you can send url encoded json 🤮
	})

	if err != nil {
		return err
	}

	resp.Body.Close()

	return nil
}

func (m *Map) createTrack(t *Track) error {
	shape := &sartopo.Shape{
		Properties: sartopo.ShapeProperties{
			Title:         t.Name,
			Description:   t.Description,
			GPSType:       "TRACK",
			StrokeWidth:   2,
			StrokeOpacity: 1,
			Stroke:        "#ff0000",
			Pattern:       "M-5 8 L0 -2 L5 8 Z,100%,,T",
			Fill:          "#ff0000",
		},
		Geometry: sartopo.ShapeGeometry{
			Coordinates: t.PointsAs2DArray(),
		},
	}

	body, err := json.Marshal(shape)
	if err != nil {
		return err
	}

	return m.post("/Shape", string(body))
}

func (m *Map) updateTrack(t *Track, id string) error {
	shape := &sartopo.Shape{
		ID:   id,
		Type: "feature",
		Properties: sartopo.ShapeProperties{
			Class:         "Shape",
			Title:         t.Name,
			Description:   t.Description,
			GPSType:       "TRACK",
			StrokeWidth:   2,
			StrokeOpacity: 1,
			Stroke:        "#ff0000",
			Pattern:       "M-5 8 L0 -2 L5 8 Z,100%,,T",
			Fill:          "#ff0000",
		},
		Geometry: sartopo.ShapeGeometry{
			Coordinates: t.PointsAs2DArray(),
		},
	}

	body, err := json.Marshal(shape)
	if err != nil {
		return err
	}

	return m.post("/Shape/"+id, string(body))
}

func (m *Map) SyncTrack(t *Track) error {
	id, ok := m.TrackIDs[t.ID]
	if !ok {
		return m.createTrack(t)
	}

	return m.updateTrack(t, id)
}

func NewMap(address string) (*Map, error) {
	split := strings.Split(address, "/m/")
	if len(split) != 2 {
		return nil, fmt.Errorf("invalid sartopo address: %s", address)
	}

	return &Map{
		SARTopoAddress: split[0],
		ID:             split[1],
		TrackIDs:       make(map[string]string),
	}, nil
}
