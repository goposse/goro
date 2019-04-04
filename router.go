// Goro
//
// Created by Yakka
// http://theyakka.com
//
// Copyright (c) 2019 Yakka LLC.
// All rights reserved.
// See the LICENSE file for licensing details and requirements.

package goro

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"
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
	return NewGroup(prefix, g.router)
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

// Add creates a new Route using the POST method and registers the instance within the Router
func (g *Group) POST(routePath string) *Route {
	return g.Add("POST", routePath)
}

// Add creates a new Route using the PUT method and registers the instance within the Router
func (g *Group) PUT(routePath string) *Route {
	return g.Add("PUT", routePath)
}


// Router is the main routing class
type Router struct {

	// ErrorHandler - generic error handler
	ErrorHandler http.Handler

	// ShouldCacheMatchedRoutes - if true then any matched routes should be cached
	// according to the path they were matched to
	ShouldCacheMatchedRoutes bool

	// alwaysUseFirstMatch - Should the route matcher use the first match regardless?
	// If set to false, the matcher will check allowed methods for an exact match and
	// try to fallback to a catch-all route if the method is not allowed.
	alwaysUseFirstMatch bool

	// methodNotAllowedIsError - Should the router fail if the route exists but the
	// mapped http methods do not match the one requested?
	methodNotAllowedIsError bool

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
		alwaysUseFirstMatch:      false,
		methodNotAllowedIsError:  true,
		errorHandlers:            map[int]http.Handler{},
		globalHandlers:           map[string]http.Handler{},
		staticLocations:          []StaticLocation{},
		filters:                  nil,
		routes:                   &Tree{},
		variables:                map[string]string{},
		cache:                    NewRouteCache(),
		debugLevel:               DebugLevelNone,
	}
	matcher := NewMatcher(router)
	matcher.FallbackToCatchAll = router.alwaysUseFirstMatch == false &&
		router.methodNotAllowedIsError == false
	router.routeMatcher = matcher

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

// SetAlwaysUseFirstMatch - Will the router always return the first match
// regardless of whether it fully meets all the criteria?
func (r *Router) SetAlwaysUseFirstMatch(alwaysUseFirst bool) {
	r.alwaysUseFirstMatch = alwaysUseFirst
	r.routeMatcher.FallbackToCatchAll = r.alwaysUseFirstMatch == false &&
		r.methodNotAllowedIsError == false
}

// SetMethodNotAllowedIsError - Will the router fail when it encounters a defined
// route that matches, but does not have a definition for the requested http method?
func (r *Router) SetMethodNotAllowedIsError(isError bool) {
	r.methodNotAllowedIsError = isError
	r.routeMatcher.FallbackToCatchAll = r.alwaysUseFirstMatch == false &&
		r.methodNotAllowedIsError == false
}

// NewMatcher returns a new matcher for the given Router
func (r *Router) NewMatcher() *Matcher {
	return NewMatcher(r)
}

// NewChain - returns a new chain with the current router attached
func (r *Router) NewChain(handlers ...ChainHandler) Chain {
	chain := NewChain(handlers...)
	chain.router = r
	return chain
}

// NewChainWithFuncs - returns a new chain with the current router attached
func (r *Router) NewChainWithFuncs(handlers ...ChainHandlerFunc) Chain {
	chain := NewChainWithFuncs(handlers...)
	chain.router = r
	return chain
}

func (r *Router) Group(prefix string) *Group {
	return NewGroup(prefix, r)
}

// Add creates a new Route and registers the instance within the Router
func (r *Router) Add(method string, routePath string) *Route {
	route := NewRoute(method, routePath)
	return r.Use(route)[0]
}

// Add creates a new Route using the GET method and registers the instance within the Router
func (r *Router) GET(routePath string) *Route {
	return r.Add("GET", routePath)
}

// Add creates a new Route using the POST method and registers the instance within the Router
func (r *Router) POST(routePath string) *Route {
	return r.Add("POST", routePath)
}

// Add creates a new Route using the PUT method and registers the instance within the Router
func (r *Router) PUT(routePath string) *Route {
	return r.Add("PUT", routePath)
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

	method := strings.ToUpper(req.Method)
	initialContext := req.Context()
	if initialContext == nil {
		initialContext = context.Background()
	}

	if r.filters != nil {
		for _, filter := range r.filters {
			originalReq := req.WithContext(initialContext)
			filter.ExecuteFilter(&originalReq)
			req = originalReq
			initialContext = req.Context() // update the working context so we can pass it along
		}
	}

	usePath := CleanPath(req.URL.Path)
	outCtx := context.WithValue(initialContext, PathContextKey, usePath)

	if r.ErrorHandler != nil {
		defer r.recoverPanic(outCtx, w, req)
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
				outCtx = context.WithValue(outCtx, ParametersContextKey, match.Params)
				if match.CatchAllValue != "" {
					outCtx = context.WithValue(outCtx, CatchAllValueContextKey, match.CatchAllValue)
				}
				useReq := req.WithContext(outCtx)
				if len(r.BeforeChain.Handlers) > 0 {
					chain := r.BeforeChain
					chain.resultCompletedFunc = func(result ChainResult) {
						resultReq := result.Request
						if result.Status == ChainCompleted {
							handler.ServeHTTP(w, resultReq)
						} else {
							// TODO - make this block of code generic
							statusCode := result.StatusCode
							if statusCode == 0 {
								statusCode = http.StatusInternalServerError
							}
							errorMessage := "Server execution failed"
							if result.Error != nil {
								errorMessage = result.Error.Error()
							}
							err := ErrorMap{
								"code":        RouterErrorCode(result.Status),
								"status_code": statusCode,
								"message":     errorMessage,
							}
							outCtx = context.WithValue(outCtx, ErrorValueContextKey, err)
							r.emitError(
								w,
								resultReq.WithContext(outCtx),
								errors.New(matchError),
								matchErrorCode,
							)
						}
					}
					chain.ServeHTTP(w, useReq)
				} else {
					handler.ServeHTTP(w, useReq)
				}
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
		err := ErrorMap{
			"code":        matchErrorCode,
			"status_code": matchErrorCode,
			"message":     matchError,
		}
		outCtx = context.WithValue(outCtx, ErrorValueContextKey, err)
		r.emitError(
			w,
			req.WithContext(outCtx),
			errors.New(matchError),
			matchErrorCode,
		)
	}
}

func (r *Router) shouldServeStaticFile(w http.ResponseWriter, req *http.Request, servePath string) (fileExists bool, filePath string) {
	if r.staticLocations != nil && len(r.staticLocations) > 0 {
		for _, staticDir := range r.staticLocations {
			seekPath := servePath
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

// error handling
func (r *Router) emitError(w http.ResponseWriter, req *http.Request, err error, errCode int) {
	// try to call specific error handler
	errHandler := r.errorHandlers[errCode]
	if errHandler != nil {
		errHandler.ServeHTTP(w, req)
		return
	}
	// if generic error handler defined, call that
	if r.ErrorHandler != nil {
		r.ErrorHandler.ServeHTTP(w, req)
		return
	}
	// return a generic http error
	errorHandler(w, req, err.Error(), errCode)

}

func errorHandler(w http.ResponseWriter, req *http.Request, errorString string, errorCode int) {
	http.Error(w, errorString, errorCode)
}

func (r *Router) recoverPanic(ctx context.Context, w http.ResponseWriter, req *http.Request) {
	if panicRecover := recover(); panicRecover != nil {
		var message string
		switch panicRecover.(type) {
		case error:
			message = panicRecover.(error).Error()
		case string:
			message = panicRecover.(string)
		default:
			message = "Panic! Please check the 'error' value for details"
		}
		err := ErrorMap{
			"code":        ErrorCodePanic,
			"status_code": http.StatusInternalServerError,
			"message":     message,
			"error":       panicRecover,
		}
		outCtx := context.WithValue(ctx, ErrorValueContextKey, err)
		r.ErrorHandler.ServeHTTP(w, req.WithContext(outCtx))
	}
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
