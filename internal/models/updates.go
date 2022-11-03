package models

import "time"

type updateRequest struct {
	timestamp time.Time
	update    string
	result    chan error
}
