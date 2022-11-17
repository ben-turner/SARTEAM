package mutationapi

import (
	"testing"
	"time"
)

var constTime = time.Now()

func TestMain(m *testing.M) {
	timeNow = func() time.Time {
		return constTime
	}

	generateMutationID = func() MutationID {
		return "test-mutation-id"
	}

	m.Run()
}
