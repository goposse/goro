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

func expectHitResult(t *testing.T, router *goro.Router, method string, path string) {
	Debug("Requesting", path, "...")
	execMockRequest(router, method, path)
	if !wasHit {
		t.Error("Expected", path, "to be HIT but it wasn't")
	}
	resetState()
}

func expectNotHitResult(t *testing.T, router *goro.Router, method string, path string) {
	Debug("Requesting", path, "...")
	execMockRequest(router, method, path)
	if wasHit {
		t.Error("Expected", path, "to be NOT HIT but it wasn't")
	}
	resetState()
}

func execMockRequest(router *goro.Router, method string, path string) {
	req, _ := http.NewRequest(method, path, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
}

func Debug(v ...interface{}) {
	if !printDebug {
		return
	}
	log.Println(v...)
}
