package sarteam

import (
	"context"
	"net/http"

	"github.com/ben-turner/sarteam/internal/models"
	"github.com/ben-turner/sarteam/internal/radiotracker"
	"golang.org/x/net/websocket"
)

type SARTeam struct {
	RootModel *models.SARTeam

	conns []*Conn

	Context      context.Context
	Cancel       context.CancelFunc
	RadioTracker *radiotracker.RadioTracker
}

func (s *SARTeam) handleRadioTracks() {
	for {
		select {
		case msg := <-s.RadioTracker.Messages():
			incident := s.RootModel.ActiveIncident
			if incident == nil {
				continue
			}

			for _, team := range incident.Teams {
				if team.State != models.AssetStateActive || !team.HasRadio(msg.RadioID) {
					continue
				}

				point := &models.Point{
					Latitude:  msg.Latitude,
					Longitude: msg.Longitude,
					Time:      msg.Timestamp,
				}
				team.RadioTracks[msg.RadioID].AddPoint(point)
			}

		case <-s.Context.Done():
			return
		}
	}
}

func (s *SARTeam) handleWS(ws *websocket.Conn) {
	conn := &Conn{
		ws: ws,
	}

	s.conns = append(s.conns, conn)
	conn.Start(s)
}

// Run starts the SARTeam application.
func (s *SARTeam) Run() error {
	go s.handleRadioTracks()

	router := http.NewServeMux()
	router.Handle("/ws", websocket.Handler(s.handleWS))
	router.Handle("/", http.FileServer(http.Dir(s.RootModel.Config.Paths.Web)))

	return http.ListenAndServe(s.RootModel.Config.ListenAddress, router)
}

func New(config *models.Config) *SARTeam {
	ctx, cancel := context.WithCancel(context.Background())

	return &SARTeam{
		RootModel: models.NewRoot(config),

		Context:      ctx,
		Cancel:       cancel,
		RadioTracker: radiotracker.New(ctx),
	}
}
