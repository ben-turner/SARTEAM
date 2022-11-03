package sarteam

import "github.com/gorilla/websocket"

type Conn struct {
	ws *websocket.Conn
}

func (c *Conn) Start(s *SARTeam) {

}
