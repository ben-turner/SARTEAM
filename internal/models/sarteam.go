package models

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/websocket"
)

var (
	ErrNoActiveIncident = errors.New("no active incident")
	ErrIncidentNotFound = errors.New("incident not found")
	ErrInvalidWorkDir   = errors.New("invalid workdir")
)

// SARTeam is the root model for the application.
type SARTeam struct {
	mux *http.ServeMux

	// A slice of all open connections.
	conns []*Conn

	netstatus       bool
	netstatusTicker *time.Ticker

	upgrader *websocket.Upgrader

	// A channel for new connections.
	ws chan *websocket.Conn

	// A channel of mutations to be applied.
	mutations chan *mutation

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

func (s *SARTeam) get(mut *mutation) error {
	t := mut.Pop(1)
	if len(t) != 1 {
		return ErrInvalidCommand
	}

	switch t[0] {
	case "NETSTATUS":
		mut.Reply(fmt.Sprintf("SET NETSTATUS %t", s.netstatus))

	default:
		return ErrInvalidCommand
	}

	return nil
}

func (s *SARTeam) set(mut *mutation) error {
	switch mut.command[0] {
	case "active":
		oldIncident := s.activeIncident
		s.activeIncident = mut.command[1]
		mut.undoFuncs = append(mut.undoFuncs, func() {
			s.activeIncident = oldIncident
		})
	default:
		return ErrInvalidCommand
	}

	return nil
}

func (s *SARTeam) incident(mut *mutation) error {
	idRes := mut.Pop(1)

	if len(idRes) != 1 {
		return ErrInvalidCommand
	}

	id := idRes[0]

	if id == "active" {
		id = s.activeIncident
	}

	incident, ok := s.incidents[id]
	if !ok {
		return ErrIncidentNotFound
	}

	return incident.applyMutation(mut)
}

func (s *SARTeam) applyMutation(mut *mutation) error {
	cmd := mut.Pop(1)
	if len(cmd) != 1 {
		return ErrInvalidCommand
	}

	switch cmd[0] {
	case "GET":
		return s.get(mut)
	case "SET":
		return s.set(mut)
	case "INCIDENT":
		return s.incident(mut)
	}

	return ErrInvalidCommand
}

func (s *SARTeam) start() {
	for {
		select {
		case ws := <-s.ws:
			log.Printf("Connection accepted")

			conn := &Conn{
				ws:        ws,
				sarteam:   s,
				messageID: 0,
			}

			go conn.Start()

			s.conns = append(s.conns, conn)
		case mut := <-s.mutations:
			err := s.applyMutation(mut)
			if err != nil {
				mut.Error(err)
			}

		case <-s.netstatusTicker.C:
			if InternetAvailable() {
				s.mutations <- &mutation{command: []string{"SET", "NETSTATUS", "true"}, timestamp: time.Now()}
			} else {
				s.mutations <- &mutation{command: []string{"SET", "NETSTATUS", "false"}, timestamp: time.Now()}
			}
		}
	}
}

func (s *SARTeam) ListenAndServe() error {
	log.Println("Listening on", s.Config.ListenAddress)
	return http.ListenAndServe(s.Config.ListenAddress, s.mux)
}

func (s *SARTeam) UpgradeRequest(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Failed to upgrade connection:", err)
		return
	}

	s.ws <- conn
}

func NewRoot(config *Config) (*SARTeam, error) {
	s := &SARTeam{
		mux:             http.NewServeMux(),
		upgrader:        &websocket.Upgrader{},
		netstatusTicker: time.NewTicker(5 * time.Second),
		ws:              make(chan *websocket.Conn, 16),
		mutations:       make(chan *mutation, 16),
		incidents:       make(map[string]*Incident),
		Config:          config,
	}

	s.mux.HandleFunc("/ws", s.UpgradeRequest)
	s.mux.Handle("/", http.FileServer(http.Dir(config.Paths.Web)))

	go s.start()

	return s, nil
}
