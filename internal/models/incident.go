package models

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	IncidentFileExtension = ".incident"
)

// IncidentFile represents the path to a file containing an incident.
type IncidentDetails struct {
	Date     time.Time
	Location string
	Training bool
}

// Name returns the name of the incident.
func (details *IncidentDetails) Name() string {
	if details.Training {
		return fmt.Sprint("Training %s %s", details.Date, details.Location)
	}

	return fmt.Sprint("%s %s", details.Date, details.Location)
}

// Filename returns the filename of the incident.
//
// This is the same as the incident's name, but with unsafe characters escaped,
// and the file extension appended.
func (details *IncidentDetails) Filename() string {
	return url.PathEscape(details.Name()) + IncidentFileExtension
}

// IncidentDetailsFromName parses the given human-readable name and returns the incident details.
func IncidentDetailsFromName(name string) (*IncidentDetails, error) {
	split := strings.Split(name, " ")
	trg := split[0] == "Training"
	if trg {
		split = split[1:]
	}

	date, err := time.Parse("2006-01-02", split[0])
	if err != nil {
		return nil, err
	}

	return &IncidentDetails{
		Date:     date,
		Location: strings.Join(split[1:], " "),
		Training: trg,
	}, nil
}

// IncidentDetailsFromFilename parses the given filename and returns the incident details.
func IncidentDetailsFromFilename(filename string) (*IncidentDetails, error) {
	if !strings.HasSuffix(filename, IncidentFileExtension) {
		return nil, fmt.Errorf("invalid file extension")
	}

	name := strings.TrimSuffix(filename, IncidentFileExtension)
	name, err := url.PathUnescape(name)
	if err != nil {
		return nil, err
	}

	return IncidentDetailsFromName(name)
}

type Incident struct {
	// The file in which this incident is stored.
	f *os.File

	// A channel to send updates to.
	updates chan *mutation

	// The date of the incident.
	Date string `json:"date"`

	// The name of the incident.
	Location string `json:"name"`

	// Whether the incident is for training.
	Training bool `json:"training"`

	// The case number associated with the incident.
	CaseNumber string `json:"caseNumber"`

	// The description of the incident.
	Description string `json:"description"`

	// The date and time the incident was created.
	CreatedAt time.Time `json:"createdAt"`

	// The SARTopo map for the incident.
	Map *Map `json:"map"`

	// The teams assigned to the incident.
	Teams []*Team `json:"teams"`
}

// set updates the value of the given field.
func (i *Incident) set(mut *mutation) error {
	opts := mut.Pop(2)
	if len(opts) != 2 {
		return ErrInvalidCommand
	}

	field := opts[0]
	value := opts[1]

	switch field {
	case "date":
		oldVal := i.Date
		i.Date = value
		mut.undoFuncs = append(mut.undoFuncs, func() {
			i.Date = oldVal
		})
	case "location":
		oldVal := i.Location
		i.Location = value
		mut.undoFuncs = append(mut.undoFuncs, func() {
			i.Location = oldVal
		})
	case "training":
		oldVal := i.Training
		i.Training = value == "true"
		mut.undoFuncs = append(mut.undoFuncs, func() {
			i.Training = oldVal
		})
	case "caseNumber":
		oldVal := i.CaseNumber
		i.CaseNumber = value
		mut.undoFuncs = append(mut.undoFuncs, func() {
			i.CaseNumber = oldVal
		})
	case "description":
		oldVal := i.Description
		i.Description = value
		mut.undoFuncs = append(mut.undoFuncs, func() {
			i.Description = oldVal
		})
	case "createdAt":
		if !i.CreatedAt.IsZero() {
			return ErrPermissionDenied
		}
		t, err := time.Parse(time.RFC3339, value)
		if err != nil {
			return err
		}
		i.CreatedAt = t
		mut.undoFuncs = append(mut.undoFuncs, func() {
			i.CreatedAt = time.Time{}
		})
	case "map":
		oldVal := i.Map
		m, err := NewMap(value)
		if err != nil {
			return err
		}
		i.Map = m
		mut.undoFuncs = append(mut.undoFuncs, func() {
			i.Map = oldVal
		})
	default:
		return fmt.Errorf("unknown field %s", field)
	}

	return nil
}

// add creates a new sub-resource and adds it to the incident.
func (i *Incident) add(mut *mutation) error {
	opts := mut.Pop(2)
	if len(opts) != 2 {
		return ErrInvalidCommand
	}

	t := opts[0]
	id := opts[1]

	switch t {
	case "team":
		i.Teams = append(i.Teams, &Team{ID: id})
	default:
		return ErrInvalidCommand
	}

	return nil
}

// del removes the sub-resource with the given ID.
func (i *Incident) del(mut *mutation) error {
	opts := mut.Pop(2)
	if len(opts) != 2 {
		return ErrInvalidCommand
	}

	t := opts[0]
	id := opts[1]

	switch t {
	case "team":
		for idx, team := range i.Teams {
			if team.ID == id {
				i.Teams = append(i.Teams[:idx], i.Teams[idx+1:]...)
				return nil
			}
		}
		return fmt.Errorf("team %s not found", id)
	default:
		return fmt.Errorf("unknown type %s", t)
	}
}

// updateSub updates the given sub-resource.
func (i *Incident) team(mut *mutation) error {
	idRes := mut.Pop(1)
	if len(idRes) != 1 {
		return ErrInvalidCommand
	}

	id := idRes[0]

	for _, team := range i.Teams {
		if team.ID == id {
			team.applyMutation(mut)
			return nil
		}
	}

	return ErrTeamNotFound
}

// clear resets the incident to its default state.
// This is generally used prior to loading an incident from disk.
func (i *Incident) clear() {
	i.Date = ""
	i.Location = ""
	i.Training = false
	i.CaseNumber = ""
	i.Description = ""
	i.CreatedAt = time.Time{}
	i.Map = nil
	i.Teams = nil
}

// reload reloads the incident from the file, applying each mutation in order.
func (i *Incident) reload() error {
	i.clear()

	scanner := bufio.NewScanner(i.f)
	for scanner.Scan() {
		line := scanner.Text()

		mut := mutationFromString(line, nil)

		err := i.applyMutation(mut)
		if err != nil {
			return err
		}
	}

	return nil
}

// update applies a mutation string to the incident.
//
// If allowReloads is false, any reload commands are ignored. This is to avoid
// infinite loops when a file contains a reload command.
func (i *Incident) applyMutation(mut *mutation) error {
	cmd := mut.Pop(1)[0]

	switch cmd {
	case "SET":
		return i.set(mut)
	case "ADD":
		return i.add(mut)
	case "DEL":
		return i.del(mut)
	case "TEAM":
		return i.team(mut)
	case "RELOAD":
		if mut.requester != nil {
			return i.reload()
		}
	case "READ":

	default:
		return ErrInvalidCommand
	}

	return nil
}

// processUpdates consumes updates from the incident's update channel.
//
// Updates are applied to the in-memory incident and then written to the file if
// they were successful.
func (i *Incident) processUpdates() {
	for mut := range i.updates {
		logMsg := mut.LogMessage()

		err := i.applyMutation(mut)
		if err != nil {
			mut.Error(err)
		}

		_, err = i.f.WriteString(logMsg)
		if err != nil {
			mut.Undo()
			mut.Error(err)
		}
	}
}

// Name returns the name of the incident.
func (i *Incident) Name() string {
	if i.Training {
		return fmt.Sprint("Training %s %s", i.Date, i.Location)
	}

	return fmt.Sprint("%s %s", i.Date, i.Location)
}
