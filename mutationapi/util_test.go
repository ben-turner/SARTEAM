package mutationapi

import (
	"testing"

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

	mut := &Mutation{}
	mutations := make(chan *Mutation)

	conn := NoopConn()
	conn.ReceivableMutation = mut

	go Pipe(ctx, conn, mutations)

	for i := 0; i < 10; i++ {
		m, ok := <-mutations
		if !ok {
			t.Fatal("Channel closed unexpectedly")
		}

		if m != mut {
			t.Fatal("Unexpected mutation received")
		}
	}

	// Error on read
	conn.ReceivableMutation = nil
	m, ok := <-mutations
	if ok {
		t.Fatal("Channel not closed")
	}

	if m != nil {
		t.Fatal("Unexpected mutation received")
	}

	// Restart after error
	conn.ReceivableMutation = mut
	mutations = make(chan *Mutation)
	go Pipe(ctx, conn, mutations)

	for i := 0; i < 10; i++ {
		m, ok := <-mutations
		if !ok {
			t.Fatal("Channel closed unexpectedly")
		}

		if m != mut {
			t.Fatal("Unexpected mutation received")
		}
	}

	// Context done
	cancel()
	m, ok = <-mutations
	if ok {
		t.Fatal("Channel not closed")
	}

	if m != nil {
		t.Fatal("Unexpected mutation received")
	}
}
