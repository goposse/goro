// Goro
//
// Created by Posse in NYC
// http://goposse.com
//
// Copyright (c) 2016 Posse Productions LLC.
// All rights reserved.
// See the LICENSE file for licensing details and requirements.

package goro

import (
	"fmt"
	"regexp"
	"strings"
)

// Node - tree node to store route information
type Node struct {
	part     string
	nodeType RouteComponentType
	regexp   *regexp.Regexp // Not used currently
	routes   map[string]*Route
	nodes    []*Node
	parent   *Node
}

// Tree - storage for routes
type Tree struct {
	nodes []*Node
}

// NewTree creates a new Tree instance
func NewTree() *Tree {
	return &Tree{
		nodes: []*Node{},
	}
}

// NewNode creates a new Node instance and appends it to the tree
func (t *Tree) NewNode(part string, parent *Node) *Node {
	nodeType := ComponentTypeFixed
	if strings.HasPrefix(part, "*") {
		nodeType = ComponentTypeCatchAll
	} else if strings.HasPrefix(part, ":") {
		nodeType = ComponentTypeWildcard
	}
	node := &Node{
		part:     part,
		nodeType: nodeType,
		regexp:   nil,
		nodes:    []*Node{},
		parent:   parent,
		routes:   nil,
	}
	if parent == nil {
		t.nodes = append(t.nodes, node)
	} else {
		parent.nodes = append(parent.nodes, node)
	}
	return node
}

// AddRouteToTree splits the route into Nodes and adds them to the tree
func (t *Tree) AddRouteToTree(route *Route, variables map[string]string) {
	path := route.PathFormat
	deslashedPath := path
	if strings.HasPrefix(deslashedPath, "/") {
		deslashedPath = path[1:len(path)]
	}
	split := strings.Split(deslashedPath, "/")

	if route.IsRoot() {
		node := t.NewNode("/", nil)
		if node.routes == nil {
			node.routes = map[string]*Route{}
		}
		node.routes[route.Method] = route
	} else {
		// check to see if we need to do any variable substitution before parsing
		// NOTE: does not support nested variables
		// fmt.Printf("[SKSEA DEBUG] split: %s\n", split)
		processedSplit := []string{}
		for _, component := range split {
			if isVariablePart(component) {

				variable := resolveVariable(component, variables, path)
				deslashedVar := variable
				if strings.HasPrefix(deslashedVar, "/") {
					deslashedVar = deslashedVar[1:len(deslashedVar)]
				}
				splitVar := strings.Split(deslashedVar, "/")
				processedSplit = append(processedSplit, splitVar...)
			} else {
				processedSplit = append(processedSplit, component)
			}
		}

		split = processedSplit

		var parentNode *Node
		var node *Node
		for _, component := range split {
			// get an existing node for this segment or attach to the tree
			node = t.nodeForExactPart(component, parentNode)
			if node == nil {
				node = t.NewNode(component, parentNode)
			}
			parentNode = node
		}
		if node.routes == nil {
			node.routes = map[string]*Route{}
		}
		node.routes[route.Method] = route
	}
}

// nodeForExactPart - finds a Node either in the top-level Tree nodes or in the
// children of a parent node (if supplied) that has an exact string match for the
// supplied 'part' value
func (t *Tree) nodeForExactPart(part string, parentNode *Node) *Node {
	var nodesToCheck []*Node
	if parentNode == nil {
		nodesToCheck = t.nodes
	} else {
		nodesToCheck = parentNode.nodes
	}
	for _, node := range nodesToCheck {
		if node.part == part {
			return node
		}
	}
	return nil
}

// HasChildren returns true if the Node has 1 or more sub-Nodes
func (node *Node) HasChildren() bool {
	return node.nodes != nil && len(node.nodes) > 0
}

// RouteForMethod returns the route that was defined for the method or nil if
// no route is defined
func (node *Node) RouteForMethod(method string) *Route {
	if node.routes != nil && len(node.routes) > 0 {
		return node.routes[strings.ToUpper(method)]
	}
	return nil
}

// part type helper functions
func isVariablePart(part string) bool {
	return strings.HasPrefix(part, "$")
}

func isWildcardPart(part string) bool {
	return strings.HasPrefix(part, ":")
}

func isCatchAllPart(part string) bool {
	return strings.HasPrefix(part, "*")
}

func containsVariable(s string) bool {
	return strings.Contains(s, "$")
}

// Returns a string with all variables resolved.
// Takes in a string which may contain variables, a map 'variables' used
// for lookup, and a string 'path' to construct panic statement if lookup fails.
func resolveVariable(component string, variables map[string]string, path string) string {
	resolved := ""
	parts := splitVariableComponent(component)
	for _, part := range parts {
		// Split parts further to handle "/"s.
		deslashedPart := strings.Split(part, "/")
		for i, dsp := range deslashedPart {
			if containsVariable(dsp) {
				lookup := lookupVariable(dsp, variables, path)
				// Lookup returned a string containing another variable.
				if containsVariable(lookup) {
					lookup = resolveVariable(lookup, variables, path)
				}
				deslashedPart[i] = lookup
			}
			// Recombine deslashed parts after they have been resolved.
		}
		resolvedPart := strings.Join(deslashedPart, "/")
		resolved = strings.Join([]string{resolved, resolvedPart}, "")
	}
	return resolved
}

// Splits a string into variables, and non-variable strings.
// For example, "foo$bar$baz" => []string{ "foo", "$bar", "$baz" }
func splitVariableComponent(s string) []string {
	separated := []string{}
	delimIndex := strings.LastIndex(s, "$")
	if delimIndex == -1 {
		// No variable components left to split
		separated = []string{s}
	} else {
		resolveNested := splitVariableComponent(s[:delimIndex])
		separated = append(resolveNested, s[delimIndex:])
	}
	return separated
}

func lookupVariable(part string, variables map[string]string, path string) string {
	lookup := variables[part]
	if lookup == "" {
		// we couldn't substitute the requested variable as there is no value definition
		panic(fmt.Sprintf("Missing variable substitution for '%s'. route='%s'", part, path))
	}
	return lookup
}

// String is the string representation of the object when printing
func (node *Node) String() string {
	return fmt.Sprintf("goro.Node # type=%d, part=%s, children=%d, routes=%v",
		node.nodeType, node.part, len(node.nodes), node.routes)
}
