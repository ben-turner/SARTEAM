package mutationapi

import (
	"log"
	"testing"
	"time"

	"context"
)

var oldIDGenerator = generateMutationID

func TestGenerateMutationID(t *testing.T) {
	t.Parallel()

	res := oldIDGenerator()
	if len(res) != 36 {
		t.Fatal("Unexpected mutation ID")
	}

	res2 := oldIDGenerator()
	if res == res2 {
		t.Fatal("Mutation IDs are not unique")
	}
}

func TestPipe(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mut := &Mutation{}
	mutations := make(chan *Mutation)

	conn := NoopConn()
	defer conn.Close()
	conn.ReceivableMutation = mut

	go Pipe(ctx, conn, mutations)

	for i := 0; i < 10; i++ {
		select {
		case m := <-mutations:
			if m != mut {
				t.Fatal("Unexpected mutation received")
			}
		case <-time.After(100 * time.Millisecond):
			t.Fatal("No mutation received")
		}
	}

	// Error on read
	conn.ReceivableMutation = nil
	err := Pipe(ctx, conn, mutations)
	if err == nil {
		t.Fatal("Expected error")
	}

	// Restart after error
	conn.ReceivableMutation = mut
	mutations = make(chan *Mutation)
	go Pipe(ctx, conn, mutations)

	for i := 0; i < 10; i++ {
		select {
		case m := <-mutations:
			if m != mut {
				t.Fatal("Unexpected mutation received")
			}
		case <-time.After(100 * time.Millisecond):
			t.Fatal("No mutation received")
		}
	}

	log.Println("here")

	// Context done
	cancel()
	select {
	case <-mutations:
		// Allow one more mutation to be received
	default:
	}

	select {
	case <-mutations:
		t.Fatal("Unexpected mutation received")
	case <-time.After(100 * time.Millisecond):
	}
}
