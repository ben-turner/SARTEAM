package mutationapi

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// allows the time package to be mocked for testing
var timeNow = time.Now

// generateMutationID generates a unique ID for a mutation. This is currently a
// UUID but this may change in the future if file sizes become too large.
//
// This is a var in order to allow tests to mock this function.
var generateMutationID = func() MutationID {
	return MutationID(uuid.New().String())
}

// Pipe receives mutations from a *Conn and sends them to a channel until the
// *Conn is closed or the provided context is done. This is a blocking function.
// If the context is done or the *Conn returns an error, this will close the
// mutations channel.
func Pipe(ctx context.Context, conn Conn, mutations chan<- *Mutation) error {
	defer close(mutations)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			mutation, err := conn.Receive()
			if err != nil {
				return err
			}

			mutations <- mutation
		}
	}
}
