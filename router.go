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
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
)

// Router is the main routing class
type Router struct {

	// errorHandler - generic error handler
	errorHandler ContextHandler

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
	//BeforeChain Chain

	// errorHandlers - map status codes to specific handlers
	errorHandlers map[int]ContextHandler

	// globalHandlers - handlers that will match all requests for an HTTP method regardless
	// of route matching
	globalHandlers map[string]ContextHandler

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
		errorHandler:             nil,
		ShouldCacheMatchedRoutes: true,
		alwaysUseFirstMatch:      false,
		methodNotAllowedIsError:  true,
		errorHandlers:            map[int]ContextHandler{},
		globalHandlers:           map[string]ContextHandler{},
		staticLocations:          []StaticLocation{},
		filters:                  nil,
		routes:                   NewTree(),
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

// SetDebugLevel - enables or disables Debug mode
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

func (r *Router) Group(prefix string) *Group {
	return NewGroup(prefix, r)
}

// Add creates a new Route and registers the instance within the Router
func (r *Router) Add(method string, routePath string) *Route {
	route := NewRoute(method, routePath)
	return r.Use(route)[0]
}

func (r *Router) AddBundle(bundle Bundle) {
	//processBundle(bundle)
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
func (r *Router) SetGlobalHandler(method string, handler ContextHandler) {
	r.globalHandlers[strings.ToUpper(method)] = handler
}

// SetRouterErrorHandler configures a ContextHandler to handle all general router errors
// (i.e.: non-network related)
func (r *Router) SetRouterErrorHandler(handler ContextHandler) {
	r.errorHandler = handler
}

// SetErrorHandler configures a ContextHandler to handle all errors for the supplied status code
func (r *Router) SetErrorHandler(statusCode int, handler ContextHandler) {
	r.errorHandlers[statusCode] = handler
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
	// create the context we're going to use for the request lifecycle
	respWriter := NewCheckedResponseWriter(w)
	hContext := NewHandlerContext(req, respWriter, r)
	if r.errorHandler != nil {
		defer r.recoverPanic(hContext)
	}
	// execute all the filters
	if r.filters != nil && len(r.filters) > 0 {
		for _, filter := range r.filters {
			filter.ExecuteBefore(hContext)
		}
	}
	// prepare the request info
	callingRequest := hContext.Request
	method := strings.ToUpper(callingRequest.Method)
	cleanPath := CleanPath(callingRequest.URL.Path)
	hContext.Path = cleanPath
	// check if there is a global handler. if so use that and be done.
	globalHandler := r.globalHandlers[method]
	if globalHandler != nil {
		globalHandler(hContext)
		return
	}
	// check to see if there is a matching route
	match := r.routeMatcher.MatchPathToRoute(method, cleanPath, callingRequest)
	if match == nil || len(match.Node.routes) == 0 {
		// check to see if there is a file match
		fileExists, filename := r.shouldServeStaticFile(respWriter, req, cleanPath)
		if fileExists {
			ServeFile(hContext, filename, http.StatusOK)
			return
		}
		// no match
		r.emitError(hContext, "Not Found", http.StatusNotFound)
		return
	}
	route := match.Node.RouteForMethod(method)
	if route == nil {
		// method not allowed
		r.emitError(hContext, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	if match.Node.nodeType == ComponentTypeCatchAll {
		// check to see if we should serve a static file at that location before falling
		// through to the catch all
		fileExists, filename := r.shouldServeStaticFile(respWriter, req, cleanPath)
		if fileExists {
			ServeFile(hContext, filename, http.StatusOK)
			return
		}
	}
	handler := route.Handler
	if handler == nil {
		r.emitError(hContext, "No Handler defined", http.StatusInternalServerError)
		return
	}
	hContext.Parameters = NewParametersWithMap(match.Params)
	if match.CatchAllValue != "" {
		hContext.CatchAllValue = match.CatchAllValue
	}
	handler(hContext)
	r.executePostFilters(hContext)
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
func (r *Router) emitError(context *HandlerContext, errMessage string, statusCode int) {
	routingError := RoutingError{
		StatusCode: statusCode,
		Message:    errMessage,
		ErrorCode:  0,
		Error:      nil,
	}
	context.Errors = append(context.Errors, routingError)
	// try to call specific error handler
	errHandler := r.errorHandlers[statusCode]
	if errHandler != nil {
		errHandler(context)
		r.executePostFilters(context)
		return
	}
	// if generic error handler defined, call that
	if r.errorHandler != nil {
		r.errorHandler(context)
		r.executePostFilters(context)
		return
	}
	// return a generic http error
	errorHandler(context.ResponseWriter, context.Request,
		errMessage, statusCode)
	r.executePostFilters(context)
}

func (r *Router) executePostFilters(ctx *HandlerContext) {
	hasDonePost := ctx.internalState[StateKeyHasExecutedPostFilters]
	if hasDonePost == nil {
		if r.filters != nil && len(r.filters) > 0 {
			for _, filter := range r.filters {
				filter.ExecuteAfter(ctx)
			}
		}
		ctx.internalState[StateKeyHasExecutedPostFilters] = true
	}
}

func errorHandler(w http.ResponseWriter, _ *http.Request, errorString string, statusCode int) {
	http.Error(w, errorString, statusCode)
}

func (r *Router) recoverPanic(handlerContext *HandlerContext) {
	if panicRecover := recover(); panicRecover != nil {
		var message string
		var err error
		switch panicRecover.(type) {
		case error:
			err = panicRecover.(error)
			message = err.Error()
		case string:
			message = panicRecover.(string)
			err = errors.New(message)
		default:
			message = "Panic! Please check the 'error' value for details"
			err = errors.New(message)
		}
		routingErr := RoutingError{
			ErrorCode:  ErrorCodePanic,
			StatusCode: http.StatusInternalServerError,
			Message:    message,
			Error:      err,
			Info: ErrorInfoMap{
				"stack": debug.Stack(),
			},
		}
		handlerContext.Errors = append(handlerContext.Errors, routingErr)
		r.errorHandler(handlerContext)
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
