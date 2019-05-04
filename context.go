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
	"net/http"
	"sync"
)

const (
	StateKeyHasExecutedPreFilters  = "_goro.skey.hasExecutedPreFilters"
	StateKeyHasExecutedPostFilters = "_goro.skey.hasExecutedPostFilters"
)

type HandlerContext struct {
	sync.RWMutex
	Request        *http.Request
	ResponseWriter http.ResponseWriter
	Parameters     *Parameters
	Meta           map[string]interface{}
	Path           string
	CatchAllValue  string
	Errors         []RoutingError
	router         *Router
	state          map[string]interface{}
	internalState  map[string]interface{}
}

func NewHandlerContext(request *http.Request, responseWriter http.ResponseWriter, router *Router) *HandlerContext {
	return &HandlerContext{
		Request:        request,
		ResponseWriter: responseWriter,
		router:         router,
		Meta:           map[string]interface{}{},
		state:          map[string]interface{}{},
		internalState:  map[string]interface{}{},
	}
}

func (hc *HandlerContext) SetState(key string, value interface{}) {
	hc.Lock()
	hc.state[key] = value
	hc.Unlock()
}

func (hc *HandlerContext) GetState(key string) interface{} {
	hc.RLock()
	state := hc.state[key]
	hc.RUnlock()
	return state
}

func (hc *HandlerContext) ClearState(key string) {
	hc.Lock()
	hc.state[key] = nil
	hc.Unlock()
}

func (hc *HandlerContext) HasError() bool {
	return len(hc.Errors) > 0
}

func (hc *HandlerContext) ErrorForStatus(status int) RoutingError {
	for _, err := range hc.Errors {
		if err.StatusCode == status {
			return err
		}
	}
	return EmptyRoutingError()
}

func (hc *HandlerContext) HasErrorForStatus(status int) bool {
	return !IsEmptyRoutingError(hc.ErrorForStatus(status))
}

func (hc *HandlerContext) FirstError() RoutingError {
	if len(hc.Errors) > 0 {
		return hc.Errors[0]
	}
	return EmptyRoutingError()
}
