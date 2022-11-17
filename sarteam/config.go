package sarteam

import "time"

type Config struct {
	// MutationBufferSize is the size of the channel used to buffer mutations
	// from connections.
	MutationBufferSize int `json:"mutationBufferSize"`
	// How often to send a ping to the client.
	PingInterval time.Duration `json:"pingInterval"`
	// How long to wait for a pong from the client before closing the connection.
	ConnectionTimeout time.Duration `json:"connectionTimeout"`
	// The directory to serve the web interface from.
	WebDir string `json:"webDir"`
	// ListenAddr is the address to listen on.
	ListenAddr string `json:"listenAddr"`
	// StateFile is the file that mutations are logged to.
	StateFile string `json:"logFile"`
}
