// Goro
//
// Created by Yakka
// http://theyakka.com
//
// Copyright (c) 2019 Yakka LLC.
// All rights reserved.
// See the LICENSE file for licensing details and requirements.

package goro

import "path"

type Group struct {
	prefix string
	router *Router
}

func NewGroup(prefix string, router *Router) *Group {
	return &Group{
		prefix: prefix,
		router: router,
	}
}

func (g *Group) Group(prefix string) *Group {
	fullPrefix := path.Join(g.prefix, prefix)
	return NewGroup(fullPrefix, g.router)
}

// Add creates a new Route and registers the instance within the Router
func (g *Group) Add(method string, routePath string) *Route {
	route := NewRoute(method, path.Join(g.prefix, routePath))
	return g.router.Use(route)[0]
}

// Add creates a new Route using the GET method and registers the instance within the Router
func (g *Group) GET(routePath string) *Route {
	return g.Add("GET", routePath)
}

// Add creates a new Route using the HEAD method and registers the instance within the Router
func (g *Group) HEAD(routePath string) *Route {
	return g.Add("HEAD", routePath)
}

// Add creates a new Route using the POST method and registers the instance within the Router
func (g *Group) POST(routePath string) *Route {
	return g.Add("POST", routePath)
}

// Add creates a new Route using the PUT method and registers the instance within the Router
func (g *Group) PUT(routePath string) *Route {
	return g.Add("PUT", routePath)
}

// Add creates a new Route using the DELETE method and registers the instance within the Router
func (g *Group) DELETE(routePath string) *Route {
	return g.Add("DELETE", routePath)
}

// Add creates a new Route using the PATCH method and registers the instance within the Router
func (g *Group) PATCH(routePath string) *Route {
	return g.Add("PATCH", routePath)
}

// Add creates a new Route using the OPTIONS method and registers the instance within the Router
func (g *Group) OPTIONS(routePath string) *Route {
	return g.Add("OPTIONS", routePath)
}
