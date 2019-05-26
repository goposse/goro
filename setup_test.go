// Goro
//
// Created by Yakka
// http://theyakka.com
//
// Copyright (c) 2019 Yakka LLC.
// All rights reserved.
// See the LICENSE file for licensing details and requirements.

package goro_test

import (
	"github.com/theyakka/goro"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

var wasHit = false
var router *goro.Router
var sum = 0

func TestMain(m *testing.M) {
	router = goro.NewRouter()

	chainHandlers := []goro.ChainHandler{chainHandler1, chainHandler2, chainHandler3}
	haltHandlers := []goro.ChainHandler{chainHandler1, chainHandler2, testHaltHandler, chainHandler3}

	router.SetStringVariable("color", "blue")
	// router tests
	router.GET("/").Handle(testHandler)
	router.GET("/users/:id").Handle(testParamsHandler)
	router.GET("/users/:id/action/:action").Handle(testParamsHandler)
	router.GET("/colors/$color").Handle(testHandler)
	// route groups
	apiGroup := router.Group("/api")
	v1Group := apiGroup.Group("/v1")
	v1Group.GET("/").Handle(testHandler)
	v1Group.POST("/").Handle(testHandler)
	v1Group.GET("/stats").Handle(testHandler)
	apiDocsGroup := v1Group.Group("/docs")
	apiDocsGroup.GET("/stats").Handle(testHandler)
	// chain tests
	router.GET("/chain/simple").Handle(goro.HC(chainHandlers...).Call())
	router.GET("/chain/then").Handle(goro.HC(chainHandlers...).Then(testThenHandler))
	router.GET("/chain/halt").Handle(goro.HC(haltHandlers...).Call())
	if printDebug {
		router.PrintRoutes()
	}
	os.Exit(m.Run())
}

func resetState() {
	wasHit = false
	sum = 0
}

func expectHitResult(t *testing.T, handler http.Handler, method string, path string) {
	Debug("Requesting", path, "...")
	execMockRequest(handler, method, path)
	if !wasHit {
		t.Error("Expected", path, "to be HIT but it wasn't")
	}
	resetState()
}

func expectNotHitResult(t *testing.T, handler http.Handler, method string, path string) {
	Debug("Requesting", path, "...")
	execMockRequest(router, method, path)
	if wasHit {
		t.Error("Expected", path, "to be NOT HIT but it wasn't")
	}
	resetState()
}

func execMockRequest(handler http.Handler, method string, url string) {
	req, _ := http.NewRequest(method, url, nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
}

func Debug(v ...interface{}) {
	if !printDebug {
		return
	}
	log.Println(v...)
}
