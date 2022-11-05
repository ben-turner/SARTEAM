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

func (details *IncidentDetails) Filename() string {
	return url.PathEscape(details.Name()) + IncidentFileExtension
}

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
func (i *Incident) set(field, value string) error {
	switch field {
	case "date":
		i.Date = value
	case "location":
		i.Location = value
	case "training":
		i.Training = value == "true"
	case "caseNumber":
		i.CaseNumber = value
	case "description":
		i.Description = value
	case "createdAt":
		if !i.CreatedAt.IsZero() {
			return fmt.Errorf("createdAt already set")
		}
		t, err := time.Parse(time.RFC3339, value)
		if err != nil {
			return err
		}
		i.CreatedAt = t
	case "map":
		m, err := NewMap(value)
		if err != nil {
			return err
		}
		i.Map = m
	default:
		return fmt.Errorf("unknown field %s", field)
	}

	return nil
}

// add creates a new sub-resource and adds it to the incident.
func (i *Incident) add(t, id string) error {
	switch t {
	case "team":
		i.Teams = append(i.Teams, &Team{ID: id})
	default:
		return fmt.Errorf("unknown type %s", t)
	}

	return nil
}

// del removes the sub-resource with the given ID.
func (i *Incident) del(t, id string) error {
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
func (i *Incident) updateSub(t, id string, update []string) error {
	switch t {
	case "team":
		for _, team := range i.Teams {
			if team.ID == id {
				team.update(update)
				return nil
			}
		}
		return fmt.Errorf("team %s not found", id)
	default:
		return fmt.Errorf("unknown type %s", t)
	}
}

// clear clears the incident of all data.
func (i *Incident) clear() {
	// TODO: Lock reads/writes to the incident.
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
		splitLine := strings.Split(line, " ")

		_, err := time.Parse(time.RFC3339, splitLine[0])
		if err != nil {
			return err
		}

		err = i.update(splitLine[1:], false)
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
func (i *Incident) update(update []string, msg ) error {
	switch update[0] {
	case "SET":
		return i.set(update[1], update[2])
	case "ADD":
		return i.add(update[1], update[2])
	case "DEL":
		return i.del(update[1], update[2])
	case "UPDATE":
		return i.updateSub(update[1], update[2], update[3:])
	case "RELOAD":
		if allowReloads {
			return i.reload()
		}
	case "READ":

	default:
		return fmt.Errorf("unknown update %s", update[0])
	}

	return nil
}

// processUpdates consumes updates from the incident's update channel.
//
// Updates are applied to the in-memory incident and then written to the file if
// they were successful.
func (i *Incident) processUpdates() {
	for update := range i.updates {
		updates := strings.Split(update.update, " ")
		err := i.update(updates, true)
		if err != nil {
			update.result <- err
		}

		logMsg := update.timestamp.Format(time.RFC3339) + " " + update.update + "\n"
		_, err = i.f.WriteString(logMsg)
		if err != nil {
			update.result <- err
		}

		close(update.result)
	}
}

// Update applies a mutation string to the incident.
func (i *Incident) Update(update string) error {
	result := make(chan error)
	i.updates <- &mutation{
		update:    update,
		result:    result,
		timestamp: time.Now(),
	}

	return <-result
}

// Reload reloads the incident from the file.
func (i *Incident) Reload() error {
	result := make(chan error)
	i.updates <- &mutation{
		update:    "RELOAD",
		result:    result,
		timestamp: time.Now(),
	}

	return <-result
}

func (i *Incident) Name() string {
	if i.Training {
		return fmt.Sprint("Training %s %s", i.Date, i.Location)
	}

	return fmt.Sprint("%s %s", i.Date, i.Location)
}
