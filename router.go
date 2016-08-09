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
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// StaticLocation is a holder for static location information
type StaticLocation struct {
	// root is the root (source) location
	root string

	// prefix is a path prefix to applied when matching
	prefix string
}

// Router is the main routing class
type Router struct {

	// ErrorHandler - generic error handler
	ErrorHandler http.Handler

	// ShouldCacheMatchedRoutes - if true then any matched routes should be cached
	// according to the path they were matched to
	ShouldCacheMatchedRoutes bool

	// BeforeChain - a Chain of handlers that will always be executed before the Route handler
	BeforeChain Chain

	// errorHandlers - map status codes to specific handlers
	errorHandlers map[int]http.Handler

	// globalHandlers - handlers that will match all requests for an HTTP method regardless
	// of route matching
	globalHandlers map[string]http.Handler

	staticLocations []StaticLocation

	// filters - registered pre-process filters
	filters []Filter

	// routeMatcher - the primary route matcher instance
	routeMatcher *Matcher

	// methodKeyedRoutes - all routes registered with the router
	routes *Tree

	// variables - unwrapped (clean) variables that have been defined
	variables map[string]string

	// cache - matched routes to path mappings
	cache *RouteCache

	// debugLevel - if enabled will output debugging information
	debugLevel DebugLevel
}

// NewRouter - creates a new default instance of the Router type
func NewRouter() *Router {
	router := &Router{
		ErrorHandler:             nil,
		ShouldCacheMatchedRoutes: true,
		errorHandlers:            map[int]http.Handler{},
		globalHandlers:           map[string]http.Handler{},
		staticLocations:          []StaticLocation{},
		filters:                  nil,
		routes:                   &Tree{},
		variables:                map[string]string{},
		cache:                    NewRouteCache(),
		debugLevel:               DebugLevelNone,
	}
	router.routeMatcher = NewMatcher(router)
	return router
}

// SetDebugLevel - enables or disables debug mode
func (r *Router) SetDebugLevel(debugLevel DebugLevel) {
	debugTimingsOn := debugLevel == DebugLevelTimings
	debugFullOn := debugLevel == DebugLevelFull
	debugOn := debugTimingsOn || debugFullOn
	onOffString := "on"
	if !debugOn {
		onOffString = "off"
	}
	Log("Debug mode is", onOffString)
	r.debugLevel = debugLevel
	r.routeMatcher.LogMatchTime = debugOn
}

// NewMatcher returns a new matcher for the given Router
func (r *Router) NewMatcher() *Matcher {
	return NewMatcher(r)
}

// Add creates a new Route and registers the instance within the Router
func (r *Router) Add(method string, path string) *Route {
	route := NewRoute(method, path)
	return r.Use(route)[0]
}

// Use registers one or more Route instances within the Router
func (r *Router) Use(routes ...*Route) []*Route {
	for _, route := range routes {
		r.routes.AddRouteToTree(route, r.variables)
	}
	return routes
}

// AddStatic registers a directory to serve static files
func (r *Router) AddStatic(staticRoot string) {
	r.AddStaticWithPrefix(staticRoot, "")
}

// AddStaticWithPrefix registers a directory to serve static files. prefix value
// will be added at matching
func (r *Router) AddStaticWithPrefix(staticRoot string, prefix string) {
	staticLocation := StaticLocation{
		root:   staticRoot,
		prefix: prefix,
	}
	r.staticLocations = append(r.staticLocations, staticLocation)
}

// SetGlobalHandler configures a ContextHandler to handle all requests for a given method
func (r *Router) SetGlobalHandler(method string, handler http.Handler) {
	r.globalHandlers[strings.ToUpper(method)] = handler
}

// SetGlobalHandlerFunc configures a ContextHandlerFunc to handle all requests for a given method
func (r *Router) SetGlobalHandlerFunc(method string, handlerFunc http.HandlerFunc) {
	r.SetGlobalHandler(method, http.Handler(handlerFunc))
}

// SetErrorHandler configures a ContextHandler to handle all errors for the supplied status code
func (r *Router) SetErrorHandler(statusCode int, handler http.Handler) {
	r.errorHandlers[statusCode] = handler
}

// SetErrorHandlerFunc configures a ContextHandlerFunc to handle all errors for the supplied status code
func (r *Router) SetErrorHandlerFunc(statusCode int, handler http.HandlerFunc) {
	r.SetErrorHandler(statusCode, http.Handler(handler))
}

// AddFilter adds a filter to the list of pre-process filters
func (r *Router) AddFilter(filter Filter) {
	r.filters = append(r.filters, filter)
}

// SetStringVariable adds a string variable value for substitution
func (r *Router) SetStringVariable(variable string, value string) {
	varname := variable
	if !strings.HasPrefix(varname, "$") {
		varname = "$" + varname
	}
	r.variables[varname] = value
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	useReq := req
	usePath := strings.ToLower(useReq.URL.Path) // always compare lower
	usePath = filepath.Clean(usePath)
	method := strings.ToUpper(req.Method)

	initialContext := req.Context()
	if initialContext == nil {
		initialContext = context.Background()
	}
	outCtx := context.WithValue(initialContext, "path", usePath)

	if r.filters != nil {
		for _, filter := range r.filters {
			originalReq := req.WithContext(outCtx)
			filter.ExecuteFilter(&originalReq)
			req = originalReq
			outCtx = req.Context() // update the working context so we can pass it along
		}
	}

	// check to see if a global handler has been registered for the method
	globalHandler := r.globalHandlers[method]
	if globalHandler != nil {
		globalHandler.ServeHTTP(w, req.WithContext(outCtx))
		return
	}

	// check to see if we have a matching route
	matchError := ""
	matchErrorCode := 0
	match := r.routeMatcher.MatchPathToRoute(method, usePath, req)
	if match != nil && len(match.Node.routes) > 0 {
		route := match.Node.RouteForMethod(method)
		if route != nil {
			if match.Node.nodeType == ComponentTypeCatchAll {
				// check to see if we should serve a static file at that location before falling
				// through to the catch all
				fileExists, filename := r.shouldServeStaticFile(w, req, usePath)
				if fileExists {
					http.ServeFile(w, req, filename)
					return
				}
			}
			handler := route.Handler
			if handler != nil {
				outCtx = context.WithValue(outCtx, "params", match.Params)
				if match.CatchAllValue != "" {
					outCtx = context.WithValue(outCtx, "catchAll", match.CatchAllValue)
				}
				handler.ServeHTTP(w, req.WithContext(outCtx))
				return
			}
		} else {
			matchError = "Method Not Allowed"
			matchErrorCode = http.StatusMethodNotAllowed
		}
	} else {
		fileExists, filename := r.shouldServeStaticFile(w, req, usePath)
		if fileExists {
			http.ServeFile(w, req, filename)
			return
		}
		matchError = "Not Found"
		matchErrorCode = http.StatusNotFound
	}

	if matchErrorCode != 0 {
		err := map[string]interface{}{
			"code":    matchErrorCode,
			"message": matchError,
		}
		outCtx = context.WithValue(outCtx, "error", err)

		// try to call specific error handler
		errHandler := r.errorHandlers[matchErrorCode]
		if errHandler != nil {
			errHandler.ServeHTTP(w, req.WithContext(outCtx))
			return
		}
		// if generic error handler defined, call that
		if r.ErrorHandler != nil {
			r.ErrorHandler.ServeHTTP(w, req.WithContext(outCtx))
			return
		}
		// return a generic http error
		errorHandler(w, req, matchError, matchErrorCode)
	}
}

func (r *Router) shouldServeStaticFile(w http.ResponseWriter, req *http.Request, path string) (fileExists bool, filePath string) {
	if r.staticLocations != nil && len(r.staticLocations) > 0 {
		for _, staticDir := range r.staticLocations {
			seekPath := path
			if staticDir.prefix != "" {
				fullPrefix := staticDir.prefix
				if !strings.HasPrefix(fullPrefix, "/") {
					fullPrefix = "/" + fullPrefix
				}
				if !strings.HasSuffix(fullPrefix, "/") {
					fullPrefix = fullPrefix + "/"
				}
				if strings.HasPrefix(seekPath, fullPrefix) {
					seekPath = strings.TrimLeft(seekPath, fullPrefix)
				}
			}
			filename := filepath.Join(staticDir.root, seekPath)
			_, statErr := os.Stat(filename)
			if statErr == nil {
				return true, filename
			}
		}
	}
	return false, ""
}

func errorHandler(w http.ResponseWriter, req *http.Request, errorString string, errorCode int) {
	http.Error(w, errorString, errorCode)
}

// PrintTreeInfo prints debugging information about all registered Routes
func (r *Router) PrintTreeInfo() {
	for _, node := range r.routes.nodes {
		fmt.Println(" - ", node)
		printSubNodes(node, 0)
	}
}

// PrintRoutes prints route registration information
func (r *Router) PrintRoutes() {
	fmt.Println("")
	nodes := r.routes.nodes
	for _, node := range nodes {
		for _, route := range node.routes {
			printRouteDebugInfo(route)
		}
		printSubRoutes(node)
	}
	fmt.Println("")
}

func printSubRoutes(node *Node) {
	if node.HasChildren() {
		for _, node := range node.nodes {
			for _, route := range node.routes {
				printRouteDebugInfo(route)
			}
			printSubRoutes(node)
		}
	}
}

func printRouteDebugInfo(route *Route) {
	desc := route.Info[RouteInfoKeyDescription]
	if desc == nil {
		desc = ""
	}
	fmt.Printf("%9s   %-50s %s\n", route.Method, route.PathFormat, desc)
}

func printSubNodes(node *Node, level int) {
	if node.HasChildren() {
		for _, subnode := range node.nodes {
			indent := ""
			for i := 0; i < level+1; i++ {
				indent += " "
			}
			indent += "-"
			fmt.Println("", indent, " ", subnode)
			if subnode.HasChildren() {
				printSubNodes(subnode, level+1)
			}
		}
	}
}
