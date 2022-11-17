package mutationapi

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"testing"
)

func TestShouldSend(t *testing.T) {
	t.Parallel()

	tests := []struct {
		filter   []string
		path     []string
		expected bool
	}{
		{
			filter:   []string{"/foo/*/*", "!/foo/bar/*", "/foo/bar/baz"},
			path:     []string{"foo", "bop", "bip"},
			expected: true,
		},
		{
			filter:   []string{"/foo/*/*", "!/foo/bar/*", "/foo/bar/baz"},
			path:     []string{"foo", "bar", "bop"},
			expected: false,
		},
		{
			filter:   []string{"/foo/*/*", "!/foo/bar/*", "/foo/bar/baz"},
			path:     []string{"foo", "bar", "baz"},
			expected: true,
		},
		{
			filter:   []string{"/**"},
			path:     []string{"foo", "bar", "baz"},
			expected: true,
		},
		{
			filter:   []string{"!/**"},
			path:     []string{"foo", "bar", "baz"},
			expected: false,
		},
		{
			filter:   []string{"/**/foo", "!/**/foo/bar"},
			path:     []string{"foo", "bar", "baz"},
			expected: false,
		},
		{
			filter:   []string{"/**/foo", "!/**/foo/bar"},
			path:     []string{"bar", "baz", "foo"},
			expected: true,
		},
		{
			filter:   []string{"/**/foo", "!/**/foo/bar"},
			path:     []string{"baz", "foo", "bar"},
			expected: false,
		},
	}

	var conn Conn = nil

	for i, test := range tests {
		log.Println(i)
		filter := NewFilterConn(conn, test.filter)
		t.Run(fmt.Sprintf("test %d", i), func(t *testing.T) {
			mut := &Mutation{Path: test.path}

			e := json.NewEncoder(os.Stdout)
			e.SetIndent("", "  ")
			// e.Encode(filter.(*filterConn).filter)

			actual := filter.(*filterConn).shouldSend(mut)
			if actual != test.expected {
				t.Errorf("expected %v, got %v", test.expected, actual)
			}
		})
	}
}
