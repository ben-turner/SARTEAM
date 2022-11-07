package sarteam

import (
	"context"
	"net/http"

	"github.com/ben-turner/sarteam/mutationapi"
	"github.com/gorilla/websocket"
)

type SARTeam struct {
	*http.ServeMux

	config *Config

	ctx context.Context

	conns     mutationapi.ConnSet
	mutations chan *mutationapi.Mutation
}

func (s *SARTeam) conn(conn mutationapi.Conn) {
	s.conns.Add(conn)
	go mutationapi.Pipe(s.ctx, conn, s.mutations)
}

func (s *SARTeam) ListenAndServe() error {
	return http.ListenAndServe(s.config.ListenAddr, s)
}

func New(config *Config) *SARTeam {
	s := &SARTeam{
		ServeMux: http.NewServeMux(),

		config: config,

		ctx: context.Background(),

		conns:     mutationapi.NewConnSet(),
		mutations: make(chan *mutationapi.Mutation, config.ConnectionBufferSize),
	}

	wsHandler := &mutationapi.WebsocketHandler{
		Handler:  mutationapi.ConnHandlerFunc(s.conn),
		Upgrader: websocket.Upgrader{},

		PingInterval: config.PingInterval,
		Timeout:      config.ConnectionTimeout,
	}

	s.Handle("/ws", wsHandler)
	s.Handle("/", http.FileServer(http.Dir(config.WebDir)))

	return s
}
