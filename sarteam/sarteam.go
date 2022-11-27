package sarteam

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/ben-turner/sarteam/mutationapi"
	"github.com/gorilla/websocket"
)

type SARTeam struct {
	*http.ServeMux

	ctx    context.Context
	cancel context.CancelFunc

	config *Config

	conns *mutationapi.ConnSet

	mutations chan *mutationapi.Mutation

	state *mutationapi.MutableState
}

// work is a blocking method that does all the work of the server. It is
// responsible for applying mutations to the server state, and broadcasting
// accepted mutations to all connected clients.
// It is a blocking method that is expected to be called exactly once.
func (s *SARTeam) work() {
	for {
		select {
		case <-s.ctx.Done():
			return
		case m := <-s.mutations:
			log.Printf("%s %s %s", m.Action, m.Path, m.Body)
			err := s.state.ApplyMutation(m)
			if err != nil {
				log.Printf("ERROR %s %s", m.ClientID, err)
				m.Error(err)
				continue
			}

			if m.Action != mutationapi.MutationActionRead {
				s.conns.Broadcast(m)
			}
		}
	}
}

// AddConn adds a new connection to the server. The server will start processing
// mutations from the connection, and will send accepted mutations to the
// connection.
func (s *SARTeam) AddConn(conn mutationapi.Conn) {
	log.Printf("CONN %s", conn.String())

	s.conns.Add(conn)

	// Does nothing if the context is already done.
	go mutationapi.Pipe(s.ctx, conn, s.mutations)
}

// ErrAlreadyRunning is returned when a server is started that is already
// running.
var ErrAlreadyRunning = errors.New("ERROR server is already running")

// Start starts the SARTeam server. It blocks until the server is stopped.
func (s *SARTeam) Start(ctx context.Context) error {
	log.Println("START starting SARTEAM server")

	s.ctx, s.cancel = context.WithCancel(ctx)

	s.conns.PipeAll(s.ctx, s.mutations)
	go s.work()

	l, err := net.Listen("tcp", s.config.ListenAddr)
	if err != nil {
		return err
	}

	go func() {
		<-s.ctx.Done()
		l.Close()
	}()

	return http.Serve(l, s)
}

// Stop stops the SARTeam server.
func (s *SARTeam) Stop() {
	log.Println("STOP Stop() called")
	s.cancel()
}

// New creates a new SARTeam server.
func New(config *Config) (*SARTeam, error) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Create a cancelled context so we can use it before starting

	state, err := CreateState(config)
	if err != nil {
		return nil, err
	}

	s := &SARTeam{
		ServeMux:  http.NewServeMux(),
		ctx:       ctx,
		config:    config,
		conns:     mutationapi.NewConnSet(),
		mutations: make(chan *mutationapi.Mutation, config.MutationBufferSize),
		state:     state,
	}

	// Set up top-level mutation log. This is used to store mutations that are
	// applied to the server state, excluding incidents which are stored in their
	// own log.
	f, err := os.OpenFile(config.StateFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	fileConn := mutationapi.NewIOConn(f, f.Name())
	filterConn := fileConn // mutationapi.NewFilterConn(fileConn, []string{"!incidents/**"})
	if err != nil {
		return nil, err
	}
	s.AddConn(filterConn)

	wsHandler := &mutationapi.WebsocketHandler{
		Handler:       s.AddConn,
		Upgrader:      websocket.Upgrader{},
		PingInterval:  config.PingInterval,
		Timeout:       config.ConnectionTimeout,
		SendQueueSize: 16, // TODO: Make configurable
		RecQueueSize:  16,
	}

	wsHandler.Handler = s.AddConn

	s.Handle("/ws", wsHandler)
	s.Handle("/", http.FileServer(http.Dir(config.WebDir)))

	return s, nil
}
