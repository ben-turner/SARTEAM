package mutationapi

import (
	"strings"
)

type filterType uint8

const (
	// filterStatusNoMatch indicates that there is no defined behavior for the
	// given path and the default behavior should be used.
	filterTypeUndefined filterType = iota
	// filterTypeInclude indicates that a filter rule matches a path and should
	// be included.
	filterTypeInclude
	// filterTypeExclude indicates that a filter rule matches a path and should
	// be excluded.
	filterTypeExclude
	// filterTypeImpliedExclude indicates that a filter rule matches a path and
	// should be excluded, but that the rule was implied by another rule.
	filterTypeImpliedExclude
)

func (t filterType) String() string {
	switch t {
	case filterTypeUndefined:
		return "undefined"
	case filterTypeInclude:
		return "include"
	case filterTypeExclude:
		return "exclude"
	case filterTypeImpliedExclude:
		return "implied exclude"
	default:
		return "unknown"
	}
}

// filterRule rule that can be used to filter mutations.
type filterRule struct {
	T        filterType
	Children map[string]*filterRule
}

// filterConn is a Conn that filters out mutations that don't match the given
// set of paths.
type filterConn struct {
	// Conn is the wrapped connection.
	Conn

	// filter is a tree representing which paths should be allowed.
	filter *filterRule
}

func searchTree(tree *filterRule, path []string) filterType {
	if tree == nil {
		return filterTypeUndefined
	}

	if len(path) == 0 {
		return tree.T
	}

	res := searchTree(tree.Children[path[0]], path[1:])

	if res == filterTypeUndefined {
		res = searchTree(tree.Children["*"], path[1:])
	}

	for i := 1; i <= len(path) &&
		(res == filterTypeUndefined ||
			res == filterTypeImpliedExclude); i++ {
		tail := path[i:]
		newRes := searchTree(tree.Children["**"], tail)
		if newRes != filterTypeUndefined {
			res = newRes
		}
	}

	return res
}

// shouldSend returns true if the given mutation should be sent to the wrapped
// connection.
//
// We first walk down the filter tree to find the most specific rule that
// matches the mutation. If no exact match is found, we walk up the tree checking
// less specific wildcard rules.
func (c *filterConn) shouldSend(mutation *Mutation) bool {
	res := searchTree(c.filter, mutation.Path)
	return res != filterTypeExclude && res != filterTypeImpliedExclude

}

// Send sends a mutation to the wrapped connection if it matches one of the
// given paths.
func (c *filterConn) Send(mutation *Mutation) error {
	if !c.shouldSend(mutation) {
		return nil
	}

	return c.Conn.Send(mutation)
}

func (c *filterConn) AddRule(rule string) {
	t := filterTypeInclude
	if rule[0] == '!' {
		t = filterTypeExclude
		rule = rule[1:]
	}

	path := strings.Split(rule, "/")

	if path[0] == "" {
		path = path[1:]
	}

	curr := c.filter

	for _, p := range path {
		// Including a named path will implicitly exclude siblings.
		if t == filterTypeInclude && p != "*" && p != "**" {
			nextChild, ok := curr.Children["*"]
			if !ok {
				nextChild = &filterRule{
					T:        filterTypeUndefined,
					Children: make(map[string]*filterRule),
				}
				curr.Children["*"] = nextChild
			}
			if nextChild.T == filterTypeUndefined {
				nextChild.T = filterTypeImpliedExclude
			}

			anyChild, ok := curr.Children["**"]
			if !ok {
				anyChild = &filterRule{
					T:        filterTypeUndefined,
					Children: make(map[string]*filterRule),
				}
				curr.Children["**"] = anyChild
			}
			if anyChild.T == filterTypeUndefined {
				anyChild.T = filterTypeImpliedExclude
			}
		}

		next, ok := curr.Children[p]
		if !ok {
			next = &filterRule{
				T:        filterTypeUndefined,
				Children: make(map[string]*filterRule),
			}
			curr.Children[p] = next
		}

		curr = next
	}

	curr.T = t
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
func NewFilterConn(conn Conn, ruleStrings []string) Conn {
	c := &filterConn{
		Conn:   conn,
		filter: &filterRule{Children: make(map[string]*filterRule)},
	}

	for _, rule := range ruleStrings {
		c.AddRule(rule)
	}

	return c
}
