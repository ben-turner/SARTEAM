package models

type Config struct {
	RadioPort string `json:"radioPort" yaml:"radioPort"`
	RadioBaud int    `json:"radioBaud" yaml:"radioBaud"`

	Paths struct {
		Incidents string `json:"incidents" yaml:"incidents"`
		Web       string `json:"web" yaml:"web"`
		ConfigLog string `json:"configLog" yaml:"configLog"`
	} `json:"paths" yaml:"paths"`

	ListenAddress string `json:"listenAddress" yaml:"listenAddress"`
}
