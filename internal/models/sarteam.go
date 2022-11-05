package models

import (
	"errors"
	"os"
	"path/filepath"
)

var (
	ErrNoActiveIncident = errors.New("no active incident")
	ErrIncidentNotFound = errors.New("incident not found")
	ErrInvalidWorkDir   = errors.New("invalid workdir")
)

// SARTeam is the root model for the application.
type SARTeam struct {
	// A channel of mutations to be applied.
	updates chan *mutation

	// A map of filepaths to incidents.
	incidents map[string]*Incident

	// The currently active incident.
	activeIncident string

	// The currently loaded configuration.
	Config *Config
}

// ListIncidents returns a list of all incidents found in the working directory.
// The resulting IncidentFiles can then be used to load the incidents.
// All errors are returned as a slice of errors.
func (s *SARTeam) ListIncidents() ([]*IncidentDetails, error) {
	files := []*IncidentDetails{}

	subpaths, err := os.ReadDir(s.Config.Paths.Incidents)
	if err != nil {
		return nil, err
	}

	for _, subpath := range subpaths {
		if subpath.IsDir() {
			continue
		}

		file, err := IncidentDetailsFromFilename(subpath.Name())
		if err != nil {
			continue
		}

		files = append(files, file)
	}

	return files, nil
}

// OpenIncident opens the incident with the given name.
func (s *SARTeam) OpenIncident(details *IncidentDetails) (*Incident, error) {
	filename := details.Filename()

	incident, ok := s.incidents[filename]
	if ok {
		return incident, nil
	}

	filepath := filepath.Join(s.Config.Paths.Incidents, filename)
	file, err := os.OpenFile(filepath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0600)
	if err != nil {
		return nil, err
	}

	incident = &Incident{
		f:       file,
		updates: make(chan *mutation),
		Teams:   make([]*Team, 0),
	}

	err = incident.reload()
	if err != nil {
		return nil, err
	}

	go incident.processUpdates()

	s.incidents[filepath] = incident

	return incident, nil
}

func (s *SARTeam) applyMutation(mut *mutation) error {
	cmd := mut.command[0]
	switch cmd {
	case "SET":
		mut.command = mut.command[1:]
		s.set(mut)
	case "INCIDENT":
		id := mut.command[1]
		if id == "active" {
			id = s.activeIncident
		}

		incident, ok := s.incidents[id]
		if !ok {
			return ErrIncidentNotFound
		}

		mut.command = mut.command[2:]
		incident.applyMutation(mut)
	}

}

func (s *SARTeam) processUpdates() {
	for mutation := range s.updates {
		cmd := mutation.command[0]
		switch cmd {
		case "SET":
			mutation.command = mutation.command[1:]
			s.set(mutation)
		case "INCIDENT":
			id := mutation.command[1]
			if id == "active" {
				id = s.activeIncident
			}

			incident, ok := s.incidents[id]
			if !ok {
				mutation.Err(ErrIncidentNotFound)
				continue
			}

			mutation.command = mutation.command[2:]
			incident.applyMutation(mutation)
		}
	}
}

func (s *SARTeam) ApplyMutation(mutation *mutation) {
	s.updates <- mutation
}

func NewRoot(config *Config) *SARTeam {
	s := &SARTeam{
		incidents: make(map[string]*Incident),
		Config:    config,
	}

	go s.processUpdates()

	return s
}
