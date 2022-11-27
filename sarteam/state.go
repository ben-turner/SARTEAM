package sarteam

import "github.com/ben-turner/sarteam/mutationapi"

type State struct {
	NetworkStatus bool `mutationapi:"networkStatus"`

	Config *Config `mutationapi:"config"`

	Incidents map[string]*Incident `mutationapi:"incidents"`
}

func CreateState(config *Config) (*mutationapi.MutableState, error) {
	s := &State{
		Config: config,
	}

	return mutationapi.NewMutableState(s)
}
