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

type HandlerContext struct {
	sync.RWMutex
	Request        *http.Request
	ResponseWriter http.ResponseWriter
	Parameters     *Parameters
	Meta           map[string]interface{}
	Path           string
	CatchAllValue  string
	Errors         []ErrorMap
	router         *Router
	state          map[string]interface{}
}

func NewHandlerContext(request *http.Request, responseWriter http.ResponseWriter, router *Router) *HandlerContext {
	return &HandlerContext{
		Request:        request,
		ResponseWriter: responseWriter,
		router:         router,
		Meta:           map[string]interface{}{},
		state:          map[string]interface{}{},
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
