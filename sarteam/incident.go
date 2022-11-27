package sarteam

import (
	"encoding/json"
	"fmt"

	"github.com/ben-turner/sarteam/mutationapi"
)

type APIIncident struct {
	Name        *string `json:"name"`
	Date        *string `json:"date"`
	Location    *string `json:"location"`
	Training    *bool   `json:"training"`
	CaseNumber  *string `json:"caseNumber"`
	Description *string `json:"description"`
	MapURL      *string `json:"map"`
}

type Incident struct {
	Date        string  `mutationapi:"date"`
	Location    string  `mutationapi:"location"`
	Training    bool    `mutationapi:"training"`
	CaseNumber  string  `mutationapi:"caseNumber"`
	Description string  `mutationapi:"description"`
	Map         *Map    `mutationapi:"map"`
	Teams       []*Team `mutationapi:"teams"`
}

func (i *Incident) Name() string {
	if i.Training {
		return fmt.Sprintf("Training %s %s", i.Date, i.Location)
	}

	return fmt.Sprintf("%s %s", i.Date, i.Location)
}

func (i *Incident) ValueToAPI() *APIIncident {
	n := i.Name()
	var m *string = nil
	if i.Map != nil {
		m = &i.Map.URL
	}

	return &APIIncident{
		Name:        &n,
		Date:        &i.Date,
		Location:    &i.Location,
		Training:    &i.Training,
		CaseNumber:  &i.CaseNumber,
		Description: &i.Description,
		MapURL:      m,
	}
}

func (i *Incident) ValueToJSON() ([]byte, error) {
	apiIncident := i.ValueToAPI()

	return json.Marshal(apiIncident)
}

func (i *Incident) ValueFromJSON(b []byte) error {
	apiIncident := &APIIncident{}

	err := json.Unmarshal(b, apiIncident)
	if err != nil {
		return err
	}

	if apiIncident.Name != nil {
		return fmt.Errorf("cannot set incident name")
	}

	if apiIncident.Date != nil {
		i.Date = *apiIncident.Date
	}

	if apiIncident.Location != nil {
		i.Location = *apiIncident.Location
	}

	if apiIncident.Training != nil {
		i.Training = *apiIncident.Training
	}

	if apiIncident.CaseNumber != nil {
		i.CaseNumber = *apiIncident.CaseNumber
	}

	if apiIncident.Description != nil {
		i.Description = *apiIncident.Description
	}

	if apiIncident.MapURL != nil {
		i.Map.URL = *apiIncident.MapURL
	}

	return nil
}

func (i *Incident) GetField(field string) (mutationapi.Mutable, error) {
	switch field {
	case "date":
		return mutationapi.MakeMutable(&i.Date)
	case "location":
		return mutationapi.MakeMutable(&i.Location)
	case "training":
		return mutationapi.MakeMutable(&i.Training)
	case "caseNumber":
		return mutationapi.MakeMutable(&i.CaseNumber)
	case "description":
		return mutationapi.MakeMutable(&i.Description)
	case "name":
		return mutationapi.MakeReadOnly(i.Name())
	default:
		return nil, fmt.Errorf("invalid field: %q", field)
	}
}
