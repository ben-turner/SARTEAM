package mutationapi

import "testing"

func TestBroadcast(t *testing.T) {
	t.Parallel()

	errs, cancel := BlackholeChan[error]()
	defer cancel()

	conn1Mutations := make(chan *Mutation, 1)
	conn2Mutations := make(chan *Mutation, 1)

	conn1 := ConnStub(conn1Mutations, errs)
	conn2 := ConnStub(conn2Mutations, errs)

	set := NewConnSet()
	set.Add(conn1)
	set.Add(conn2)

	mut := &Mutation{}
	set.Broadcast(mut)

	select {
	case m1, ok := <-conn1Mutations:
		if !ok {
			t.Fatal("Channel closed unexpectedly")
		}

		if m1 != mut {
			t.Fatal("Unexpected mutation received")
		}
	default:
		t.Fatal("No mutation received")
	}

	select {
	case m2, ok := <-conn2Mutations:
		if !ok {
			t.Fatal("Channel closed unexpectedly")
		}

		if m2 != mut {
			t.Fatal("Unexpected mutation received")
		}
	default:
		t.Fatal("No mutation received")
	}
}

func TestConnSet(t *testing.T) {
	t.Parallel()

	conn1 := NoopConn()
	conn2 := NoopConn()

	set := NewConnSet()
	set.Add(conn1)
	set.Add(conn2)

	if len(set) != 2 {
		t.Fatal("Unexpected length")
	}

	set.Remove(conn1)
	if len(set) != 1 {
		t.Fatal("Unexpected length")
	}

	set.Remove(conn2)
	if len(set) != 0 {
		t.Fatal("Unexpected length")
	}

	set.Add(conn1)
	set.Add(conn2)

	conn1.Close()
	set.Purge()
	if len(set) != 1 {
		t.Fatal("Unexpected length")
	}
}
