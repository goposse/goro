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
)

// ChainStatus - the status of the chain
type ChainStatus int

const (
	// ChainCompleted - the chain completed normally
	ChainCompleted ChainStatus = 1 << iota
	// ChainError - the chain was stopped because of an error
	ChainError
	// ChainHalted - the chain was halted before it could finish executing
	ChainHalted
)

// ChainResult - the chain execution result
type ChainResult struct {
	Status     ChainStatus
	Error      error
	StatusCode int
}

type ChainHandler func(*Chain, *HandlerContext)

// ChainCompletedFunc - callback function executed when chain execution has
// completed
type ChainCompletedFunc func(result ChainResult)

// Chain allows for chaining of Handlers
type Chain struct {
	handlerIndex int
	router       *Router

	// RouterCatchesErrors - if true and the chain is attached to a router then
	// errors will bubble up to the router error handler
	RouterCatchesErrors bool

	// EmitHTTPError - if true, the router will emit an http.Error when the chain
	// result is an error
	EmitHTTPError bool

	// Handlers - the handlers in the Chain
	handlers []ChainHandler

	completedCallback ChainCompletedFunc

	// ChainCompletedFunc - called when chain completes
	ChainCompletedFunc ChainCompletedFunc
}

// NewChain - creates a new Chain instance
func NewChain(handlers ...ChainHandler) Chain {
	return Chain{
		RouterCatchesErrors: true,
		EmitHTTPError:       true,
		handlers:            handlers,
		handlerIndex:        0,
	}
}

func NewChainCopy(ch Chain) Chain {
	return copyChain(ch)
}

func HC(handlers ...ChainHandler) Chain {
	return Chain{
		RouterCatchesErrors: true,
		EmitHTTPError:       true,
		handlers:            handlers,
		handlerIndex:        0,
	}
}

// Append - returns a new chain with the ChainHandler appended to
// the list of handlers
func (ch *Chain) Append(handlers ...ChainHandler) Chain {
	allHandlers := make([]ChainHandler, 0, len(ch.handlers)+len(handlers))
	allHandlers = append(allHandlers, ch.handlers...)
	allHandlers = append(allHandlers, handlers...)
	newChain := copyChain(*ch)
	newChain.handlers = allHandlers
	return newChain
}

// Then - calls the chain and then the designated Handler
func (ch Chain) Then(handler ContextHandler) ContextHandler {
	return func(ctx *HandlerContext) {
		ch.completedCallback = func(result ChainResult) {
			handler(ctx)
		}
		ch.startChain(ctx)
	}
}

// Call - calls the chain
func (ch Chain) Call() ContextHandler {
	return func(ctx *HandlerContext) {
		ch.startChain(ctx)
	}
}

func (ch *Chain) startChain(ctx *HandlerContext) {
	ch.resetState()
	ch.handlers[0](ch, ctx)
}

func (ch *Chain) doNext(ctx *HandlerContext) {
	ch.handlerIndex++
	handlersCount := len(ch.handlers)
	if ch.handlerIndex >= handlersCount {
		// nothing to execute. notify that the chain has finished
		finish(ch, ChainCompleted, nil, 0)
		return
	}
	// execute the current chain handler
	ch.handlers[ch.handlerIndex](ch, ctx)
}

// Next - execute the next handler in the chain
func (ch *Chain) Next(ctx *HandlerContext) {
	ch.doNext(ctx)
}

// Halt - halt chain execution
func (ch *Chain) Halt(ctx *HandlerContext) {
	finish(ch, ChainHalted, nil, 0)
}

// Error - halt the chain and report an error
func (ch *Chain) Error(ctx *HandlerContext, chainError error, statusCode int) {
	finish(ch, ChainError, chainError, statusCode)
	if ch.router != nil && ch.RouterCatchesErrors {
		ch.router.emitError(ctx, chainError.Error(), statusCode)
	} else if ch.EmitHTTPError {
		http.Error(ctx.ResponseWriter, chainError.Error(), statusCode)
	}
}

func (ch Chain) Copy() Chain {
	return copyChain(ch)
}

// reset - resets the chain
func (ch *Chain) resetState() {
	ch.handlerIndex = 0
}

func finish(chain *Chain, status ChainStatus, chainError error, statusCode int) ChainResult {
	result := ChainResult{
		Status:     status,
		Error:      chainError,
		StatusCode: statusCode,
	}
	if chain.completedCallback != nil {
		chain.completedCallback(result)
	}
	if chain.ChainCompletedFunc != nil {
		chain.ChainCompletedFunc(result)
	}
	chain.resetState()
	return result
}

func copyChain(chain Chain) Chain {
	return Chain{
		RouterCatchesErrors: chain.RouterCatchesErrors,
		EmitHTTPError:       chain.EmitHTTPError,
		router:              chain.router,
		handlers:            chain.handlers,
		handlerIndex:        chain.handlerIndex,
		ChainCompletedFunc:  chain.ChainCompletedFunc,
		completedCallback:   chain.completedCallback,
	}
}
