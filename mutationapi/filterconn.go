package mutationapi

import (
	"strings"
)

type filterStatus uint8

const (
	// filterStatusNoMatch indicates that a filter rule doesn't match a path.
	filterStatusNoMatch filterStatus = iota
	// filterStatusAllow indicates that a filter rule matches a path and is not
	// inverted.
	filterStatusAllow
	// filterStatusReject indicates that a filter rule matches a path and is
	// inverted.
	filterStatusReject
)

// filterRule rule that can be used to filter mutations.
type filterRule struct {
	inverted bool
	path     []string
}

// match evaluates the given path against the filter rule. Returns
// filterStatusNoMatch if the path doesn't match the rule's path,
// filterStatusAllow if the path matches the rule's path and the rule is not
// inverted, and filterStatusReject if the path matches the rule's path and the
// rule is inverted.
func (r *filterRule) match(path []string) filterStatus {
	if len(path) < len(r.path) {
		return filterStatusNoMatch
	}

	for i, p := range r.path {
		if p != path[i] {
			return filterStatusNoMatch
		}
	}

	if r.inverted {
		return filterStatusReject
	}

	return filterStatusAllow
}

// parseFilterRule parses a filter rule string and returns a filterRule. Rules
// are a slash-separated path with an optional ! prefix.
//
// Returns ErrInvalidFilterRule for empty or invalid rules.
func parseFilterRule(rule string) (*filterRule, error) {
	if len(rule) == 0 {
		return nil, ErrInvalidFilterRule
	}

	inverted := false
	if rule[0] == '!' {
		inverted = true
		rule = rule[1:]
	}

	if len(rule) == 0 {
		return nil, ErrInvalidFilterRule
	}

	return &filterRule{
		inverted: inverted,
		path:     strings.Split(rule, "/"),
	}, nil
}

// filterConn is a Conn that filters out mutations that don't match the given
// set of paths.
type filterConn struct {
	// Conn is the wrapped connection.
	Conn

	// rules is the set of rules to match mutations against.
	rules []*filterRule
}

// shouldSend returns true if the given mutation should be sent to the wrapped
// connection.
func (c *filterConn) shouldSend(mutation *Mutation) bool {
	s := filterStatusNoMatch
	for _, rule := range c.rules {
		switch rule.match(mutation.Path) {
		case filterStatusAllow:
			s = filterStatusAllow
		case filterStatusReject:
			s = filterStatusReject
		}
	}

	// If no rule matched, the default is to reject. Therefore allow is the only
	// truthy status.
	return s == filterStatusAllow
}

// Send sends a mutation to the wrapped connection if it matches one of the
// given paths.
func (c *filterConn) Send(mutation *Mutation) error {
	if !c.shouldSend(mutation) {
		return nil
	}

	return c.Conn.Send(mutation)
}

// NewFilterConn wraps an existing Conn and filters out mutations that don't
// match the given rules.
//
// Rules are similar to the ones used in .gitignore files. A rule can be
// inverted by prefixing it with a !. Rules are evaluated in the order they are
// given. If a mutation matches multiple rules, the last one wins.
//
// Examples:
//  - /foo
//  - !/foo/bar
//  - /foo/bar/baz
//
// The above rules will allow mutations to /foo and /foo/bar/baz, but not to
// /foo/bar.
func NewFilterConn(conn Conn, ruleStrings []string) (Conn, error) {
	rules := make([]*filterRule, len(ruleStrings))
	for i, path := range ruleStrings {
		var err error
		rules[i], err = parseFilterRule(path)
		if err != nil {
			return nil, err
		}
	}

	return &filterConn{
		Conn:  conn,
		rules: rules,
	}, nil
}
