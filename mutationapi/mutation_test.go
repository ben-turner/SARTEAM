package mutationapi

import (
	"errors"
	"testing"
	"time"
)

func TestMutationActionString(t *testing.T) {
	t.Parallel()

	if MutationActionCreate.String() != "CREATE" {
		t.Errorf("expected CREATE, got %s", MutationActionCreate.String())
	}

	if MutationActionRead.String() != "READ" {
		t.Errorf("expected READ, got %s", MutationActionRead.String())
	}

	if MutationActionUpdate.String() != "UPDATE" {
		t.Errorf("expected UPDATE, got %s", MutationActionUpdate.String())
	}

	if MutationActionDelete.String() != "DELETE" {
		t.Errorf("expected DELETE, got %s", MutationActionDelete.String())
	}

	if MutationAction(0xff).String() != "UNKNOWN" {
		t.Errorf("expected UNKNOWN, got %s", MutationAction(0xFF).String())
	}
}

func TestParseMutationAction(t *testing.T) {
	t.Parallel()

	a := ParseMutationAction("CREATE")
	if a != MutationActionCreate {
		t.Fatalf("expected %v, got %v", MutationActionCreate, a)
	}

	a = ParseMutationAction("READ")
	if a != MutationActionRead {
		t.Fatalf("expected %v, got %v", MutationActionRead, a)
	}

	a = ParseMutationAction("UPDATE")
	if a != MutationActionUpdate {
		t.Fatalf("expected %v, got %v", MutationActionUpdate, a)
	}

	a = ParseMutationAction("DELETE")
	if a != MutationActionDelete {
		t.Fatalf("expected %v, got %v", MutationActionDelete, a)
	}

	a = ParseMutationAction("FOO")
	if a != MutationActionUnknown {
		t.Fatalf("expected %v, got %v", MutationActionUnknown, a)
	}
}

func TestBodyAsJSON(t *testing.T) {
	t.Parallel()

	a := &Mutation{
		body: []byte("{\"foo\": \"bar\"}"),
	}

	var m map[string]interface{}
	err := a.BodyAsJSON(&m)

	if err != nil {
		t.Fatal(err)
	}

	if m["foo"] != "bar" {
		t.Fatalf("expected bar, got %v", m["foo"])
	}

	// Fails for invalid JSON
	a.body = []byte("Hello World!")
	err = a.BodyAsJSON(&m)

	if err == nil {
		t.Fatal("expected error")
	}
}

func TestBodyAsString(t *testing.T) {
	t.Parallel()

	a := &Mutation{
		body: []byte("{\"foo\": \"bar\"}"),
	}

	if a.BodyAsString() != "{\"foo\": \"bar\"}" {
		t.Error("BodyAsString() returned unexpected result")
	}

	// Not just JSON
	a.body = []byte("Hello World!")

	if a.BodyAsString() != "Hello World!" {
		t.Error("BodyAsString() returned unexpected result")
	}
}

func TestBodyAsBool(t *testing.T) {
	t.Parallel()

	conn := NoopConn()
	defer conn.Close()

	a := &Mutation{
		ID:        "12345",
		Timestamp: time.Now(),
		Conn:      nil,
		Action:    MutationActionCreate,
		Path:      []string{"foo", "bar", "baz"},
		body:      []byte("true"),
	}

	if !a.BodyAsBool() {
		t.Fatal("expected body to be true")
	}

	a.body = []byte("false")
	if a.BodyAsBool() {
		t.Fatal("expected body to be false")
	}

	a.body = []byte("foo")
	if a.BodyAsBool() {
		t.Fatal("expected body to be false")
	}
}

func TestError(t *testing.T) {
	t.Parallel()

	sendChan, cancel := BlackholeChan[*Mutation]()
	defer cancel()

	errorChan := make(chan error, 1)

	conn := ConnStub(sendChan, errorChan)

	a := &Mutation{
		ID:        "12345",
		Timestamp: time.Now(),
		Conn:      conn,
		Action:    MutationActionCreate,
		Path:      []string{"foo", "bar", "baz"},
		body:      []byte("{\"foo\": \"bar\"}"),
	}

	sentErr := errors.New("test error")
	a.Error(sentErr)

	recErr := <-errorChan

	if recErr != sentErr {
		t.Fatalf("expected %v, got %v", sentErr, recErr)
	}
}

func TestString(t *testing.T) {
	t.Parallel()

	conn := NoopConn()
	defer conn.Close()

	a := &Mutation{
		ID:        "12345",
		Timestamp: time.Date(2022, 11, 7, 14, 5, 10, 0, time.UTC),
		Conn:      conn,
		Action:    MutationActionCreate,
		Path:      []string{"foo", "bar", "baz"},
		body:      []byte("{\"foo\": \"bar\"}"),
	}

	expected := "2022-11-07T14:05:10Z 12345 CREATE foo/bar/baz {\"foo\": \"bar\"}"

	if a.String() != expected {
		t.Fatalf("expected %v, got %v", expected, a.String())
	}

	// Sanity check that a full mutation with timestamp specified is unchanged
	// when parsed and stringified.
	msg := "2022-11-07T12:34:56Z foobar DELETE foo/bar"
	a, err := ParseMutation(msg, conn)

	if err != nil {
		t.Fatal(err)
	}

	if a.String() != msg {
		t.Fatalf("expected %q, got %q", msg, a.String())
	}

	// Check that time is converted to UTC
	a = &Mutation{
		ID:           "12345",
		Timestamp:    time.Date(2022, 11, 7, 14, 5, 10, 0, time.FixedZone("UTC-8", -8*60*60)),
		Conn:         conn,
		Action:       MutationActionCreate,
		Path:         []string{"foo", "bar", "baz"},
		body:         []byte("{\"foo\": \"bar\"}"),
		OriginalBody: []byte("{\"foo\": \"baz\"}"),
	}

	expected = "2022-11-07T22:05:10Z 12345 CREATE foo/bar/baz {\"foo\": \"bar\"}"

	if a.String() != expected {
		t.Fatalf("expected %v, got %v", expected, a.String())
	}
}

func TestInvertID(t *testing.T) {
	t.Parallel()

	a := MutationID("12345")
	b := invertID(a)

	if b != "-12345" {
		t.Fatalf("expected -12345, got %v", b)
	}

	a = MutationID("-12345")
	b = invertID(a)

	if b != "12345" {
		t.Fatalf("expected 12345, got %v", b)
	}

	a = MutationID("foobar")
	b = invertID(a)

	if b != "-foobar" {
		t.Fatalf("expected -foobar, got %v", b)
	}

	a = MutationID("-foobar")
	b = invertID(a)

	if b != "foobar" {
		t.Fatalf("expected foobar, got %v", b)
	}
}

func TestInverse(t *testing.T) {
	t.Parallel()

	conn := NoopConn()
	defer conn.Close()

	// Create/Delete
	a := &Mutation{
		ID:        "12345",
		Timestamp: constTime,
		Conn:      conn,
		Action:    MutationActionCreate,
		Path:      []string{"foo", "bar", "baz"},
		body:      []byte("{\"foo\": \"bar\"}"),
	}
	expected := &Mutation{
		ID:           "-12345",
		Timestamp:    constTime,
		Conn:         conn,
		Action:       MutationActionDelete,
		Path:         []string{"foo", "bar", "baz"},
		body:         nil,
		OriginalBody: []byte("{\"foo\": \"bar\"}"),
	}

	b := a.Inverse()
	if !b.Equal(expected) {
		t.Fatal("expected mutations to be equal")
	}

	b = b.Inverse()
	if !b.Equal(a) {
		t.Fatal("expected mutations to be equal")
	}

	// Update
	a = &Mutation{
		ID:           "12345",
		Timestamp:    constTime,
		Conn:         conn,
		Action:       MutationActionUpdate,
		Path:         []string{"foo", "bar", "baz"},
		body:         []byte("{\"foo\": \"baz\"}"),
		OriginalBody: []byte("{\"foo\": \"bar\"}"),
	}
	expected = &Mutation{
		ID:           "-12345",
		Timestamp:    constTime,
		Conn:         conn,
		Action:       MutationActionUpdate,
		Path:         []string{"foo", "bar", "baz"},
		body:         []byte("{\"foo\": \"bar\"}"),
		OriginalBody: []byte("{\"foo\": \"baz\"}"),
	}

	b = a.Inverse()
	if !b.Equal(expected) {
		t.Fatal("expected mutations to be equal")
	}

	b = b.Inverse()
	if !b.Equal(a) {
		t.Fatal("expected mutations to be equal")
	}

	// Check that the time is updated.
	a.Timestamp = constTime.Add(-10 * time.Second)
	b = a.Inverse()
	if !b.Timestamp.Equal(constTime) {
		t.Fatal("expected timestamp to be updated")
	}

	// Test invalid mutations.
	a = &Mutation{
		ID:        "12345",
		Timestamp: constTime,
		Conn:      conn,
		Action:    MutationActionUnknown,
		Path:      []string{"foo", "bar", "baz"},
		body:      []byte("{\"foo\": \"bar\"}"),
	}

	b = a.Inverse()
	if b != nil {
		t.Fatal("expected inverse to be nil")
	}
}

func TestParseMutationWith(t *testing.T) {
	t.Parallel()

	conn := NoopConn()
	defer conn.Close()

	// Test with a valid mutation with no body.
	msg := "2022-11-07T14:05:10Z 12345 READ foo/bars/baz"
	expected := &Mutation{
		ID:        "12345",
		Timestamp: time.Date(2022, 11, 7, 14, 5, 10, 0, time.UTC),
		Conn:      conn,
		Action:    MutationActionRead,
		Path:      []string{"foo", "bars", "baz"},
		body:      []byte{},
	}

	mut, err := ParseMutation(msg, conn)
	if err != nil {
		t.Fatal(err)
	}

	if !mut.Equal(expected) {
		t.Fatalf("expected %v, got %v", expected, mut)
	}

	// Test with a valid mutation with a body.
	msg = "2022-11-07T14:05:10Z 12345 CREATE foo/bars/baz {\"foo\": \"bar\"}"
	expected = &Mutation{
		ID:        "12345",
		Timestamp: time.Date(2022, 11, 7, 14, 5, 10, 0, time.UTC),
		Conn:      conn,
		Action:    MutationActionCreate,
		Path:      []string{"foo", "bars", "baz"},
		body:      []byte("{\"foo\": \"bar\"}"),
	}

	mut, err = ParseMutation(msg, conn)
	if err != nil {
		t.Fatal(err)
	}

	if !mut.Equal(expected) {
		t.Fatalf("expected %v, got %v", expected, mut)
	}

	// Test with an empty message.
	msg = ""
	_, err = ParseMutation(msg, conn)
	if err == nil {
		t.Fatal("expected error")
	}

	// Test with a message with an invalid message.
	msg = "1234 CREATE"
	_, err = ParseMutation(msg, conn)
	if err == nil {
		t.Fatal("expected error")
	}

	// Test with a message without a timestamp.
	msg = "12345 CREATE foo/bars/baz"
	expected = &Mutation{
		ID:        "12345",
		Timestamp: constTime,
		Conn:      conn,
		Action:    MutationActionCreate,
		Path:      []string{"foo", "bars", "baz"},
	}

	mut, err = ParseMutation(msg, conn)
	if err != nil {
		t.Fatal(err)
	}

	if !mut.Equal(expected) {
		t.Fatalf("expected %v, got %v", expected, mut)
	}
}

func TestEquivalent(t *testing.T) {
	t.Parallel()

	conn := NoopConn()
	defer conn.Close()

	a := &Mutation{
		ID:        "12345",
		Timestamp: time.Now(),
		Conn:      nil,
		Action:    MutationActionCreate,
		Path:      []string{"foo", "bar", "baz"},
		body:      []byte("{\"foo\": \"bar\"}"),
	}

	b := &Mutation{
		ID:        "12345",
		Timestamp: time.Now(),
		Conn:      nil,
		Action:    MutationActionCreate,
		Path:      []string{"foo", "bar", "baz"},
		body:      []byte("{\"foo\": \"bar\"}"),
	}

	if !a.Equivalent(b) {
		t.Fatal("expected mutations to be equivalent")
	}

	b.ID = "54321"
	if !a.Equivalent(b) {
		t.Fatal("expected mutations to be equivalent")
	}

	b.Conn = conn
	if !a.Equivalent(b) {
		t.Fatal("expected mutations to be equivalent")
	}

	b.OriginalBody = []byte("{\"foo\": \"bar\"}")
	if !a.Equivalent(b) {
		t.Fatal("expected mutations to be equivalent")
	}

	b.Action = MutationActionUpdate
	if a.Equivalent(b) {
		t.Fatal("expected mutations to not be equivalent")
	}

	b.Action = MutationActionCreate
	b.Path = []string{"foo", "bar", "baz", "qux"}
	if a.Equivalent(b) {
		t.Fatal("expected mutations to not be equivalent")
	}

	b.Path = []string{"foo", "bar", "oops"}
	if a.Equivalent(b) {
		t.Fatal("expected mutations to not be equivalent")
	}

	b.Path = []string{"foo", "bar", "baz"}
	b.body = []byte("{\"foo\": \"baz\"}")
	if a.Equivalent(b) {
		t.Fatal("expected mutations to not be equivalent")
	}

	// Nil arg
	if a.Equivalent(nil) {
		t.Fatal("expected mutations to not be equal")
	}
}

func TestEqual(t *testing.T) {
	t.Parallel()

	conn := NoopConn()
	defer conn.Close()

	a := &Mutation{
		ID:           "12345",
		Timestamp:    time.Now(),
		Conn:         conn,
		Action:       MutationActionCreate,
		Path:         []string{"foo", "bar", "baz"},
		body:         []byte("{\"foo\": \"bar\"}"),
		OriginalBody: []byte("{\"foo\": \"bar\"}"),
	}
	c := *a
	b := &c

	if !a.Equal(b) {
		t.Fatal("expected mutations to be equal")
	}

	// ID Mismatch
	b.ID = "54321"
	if a.Equal(b) {
		t.Fatal("expected mutations to not be equal")
	}

	// Connection Mismatch
	b.ID = "12345"
	b.Conn = nil
	if a.Equal(b) {
		t.Fatal("expected mutations to not be equal")
	}

	// OriginalBody Mismatch
	b.Conn = conn
	b.OriginalBody = []byte("{\"foo\": \"baz\"}")
	if a.Equal(b) {
		t.Fatal("expected mutations to not be equal")
	}

	// Action Mismatch
	b.OriginalBody = []byte("{\"foo\": \"bar\"}")
	b.Action = MutationActionUpdate
	if a.Equal(b) {
		t.Fatal("expected mutations to not be equal")
	}

	// Path Mismatch Length
	b.Action = MutationActionCreate
	b.Path = []string{"foo", "bar", "baz", "qux"}
	if a.Equal(b) {
		t.Fatal("expected mutations to not be equal")
	}

	// Path Mismatch Values
	b.Path = []string{"foo", "bar", "oops"}
	if a.Equal(b) {
		t.Fatal("expected mutations to not be equal")
	}

	// Body Mismatch
	b.Path = []string{"foo", "bar", "baz"}
	b.body = []byte("{\"foo\": \"baz\"}")
	if a.Equal(b) {
		t.Fatal("expected mutations to not be equal")
	}

	// Timestamp Mismatch
	b.body = []byte("{\"foo\": \"bar\"}")
	b.Timestamp = time.Now().Add(time.Second)
	if a.Equal(b) {
		t.Fatal("expected mutations to not be equal")
	}

	// Nil arg
	if a.Equal(nil) {
		t.Fatal("expected mutations to not be equal")
	}
}
