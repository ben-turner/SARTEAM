package models

type Team struct {
	// The name of the team.
	Name string `json:"name"`

	// The ID of the team.
	ID string `json:"id"`

	// The team's current task description.
	Tasking string `json:"tasking"`

	// The team's current status.
	State AssetState `json:"state"`

	// The team's leader.
	Leader *Person `json:"leader"`

	// The assistant team leader.
	ATL *Person `json:"atl"`

	// The team's members.
	Members []*Person `json:"members"`

	// The IDs of the GPS-tracked radios the team is using.
	RadioIDs []string `json:"radioId"`

	// RadioTracks is a mapping of radio IDs to the tracks they are following.
	RadioTracks map[string]*Track `json:"radioTracks"`
}

func (t *Team) HasRadio(radioID string) bool {
	for _, id := range t.RadioIDs {
		if id == radioID {
			return true
		}
	}

	return false
}

func (t *Team) update(args []string) error {
	return nil
}
