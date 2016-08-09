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
	"net/http/httptest"
	"reflect"
	"testing"
)

type RequestMock struct {
	URL                string
	Method             string
	CheckSuccess       bool
	CheckParams        bool
	CheckParamsValue   map[string][]string
	CheckCatchAll      bool
	CheckCatchAllValue string
}

type TestFilter struct {
}

func testHandler1(rw http.ResponseWriter, req *http.Request) {
}

func testHandler2(rw http.ResponseWriter, req *http.Request) {
}

func testHandler3(rw http.ResponseWriter, req *http.Request) {
}

func (tf TestFilter) ExecuteFilter(req **http.Request) {
	oldReq := *req
	newCtx := context.WithValue(oldReq.Context(), "TESTVAL", "this is a test")
	*req = oldReq.WithContext(newCtx)
}

func okHandler(rw http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	Log("OK!")
	fmt.Fprintf(rw, "OK: called '%s' -> %s", ctx.Value("path"), req.Method)
}

func errHandler(rw http.ResponseWriter, req *http.Request) {
	fmt.Fprint(rw, "error")
}

func compareMaps(map1 map[string][]string, map2 map[string][]string) bool {
	return reflect.DeepEqual(map1, map2)
}

func TestMain(t *testing.T) {

	domains := NewDomainMap()
	router := domains.NewRouter("^(?:www+[.])*(localhost.local)(?::\\d+)?")
	router.SetDebugLevel(DebugLevelTimings)

	testFilter := TestFilter{}
	router.AddFilter(testFilter)

	router.AddStaticWithPrefix("./assets", "assets")
	router.AddStaticWithPrefix("./test", "test")

	calledTestHandler := false
	var finalParams map[string][]string
	var finalCatchAll string
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		calledTestHandler = true
		ctx := req.Context()
		finalParams = ctx.Value("params").(map[string][]string)
		catchAllObj := ctx.Value("catchAll")
		finalCatchAll = ""
		if catchAllObj != nil {
			finalCatchAll = catchAllObj.(string)
		}

	})

	chain := NewChain()
	chain.AddFunc(testHandler1, testHandler3, testHandler2)

	testHandle := chain.ThenFunc(testHandler)

	router.SetStringVariable("a", "alpha$b$c")
	router.SetStringVariable("b", "bar")
	router.SetStringVariable("c", "$d")
	router.SetStringVariable("d", "baz")

	router.Add("GET", "/").
		Handle(testHandle)
	router.Add("GET", "/users/:id/*").
		Handle(testHandle)
	router.Add("GET", "/users/:id/show").
		Handle(testHandle)
	router.Add("POST", "/users/:id/show").
		Handle(testHandle)
	router.Add("GET", "/users/:id/:action").
		Handle(testHandle)
	router.Add("GET", "/users/:id/show/:what").
		Handle(testHandle)
	router.Add("GET", "/*").
		Handle(testHandle)
	router.Add("GET", "/$a$b").
		Handle(testHandle)

	reqMocks := []RequestMock{
		RequestMock{Method: "GET", URL: "/", CheckSuccess: true},
		RequestMock{Method: "POST", URL: "/", CheckSuccess: false},
		RequestMock{Method: "GET", URL: "/users/123/show", CheckSuccess: true},
		RequestMock{Method: "GET", URL: "/test/route", CheckSuccess: true},
		RequestMock{Method: "POST", URL: "/test/route", CheckSuccess: false},
		RequestMock{Method: "GET", URL: "/users/123/show/something", CheckSuccess: true},
		RequestMock{Method: "GET", URL: "users/123/show/something", CheckSuccess: true},
	}

	for _, mock := range reqMocks {
		r, _ := http.NewRequest(mock.Method, mock.URL, nil)
		w := httptest.NewRecorder()
		calledTestHandler = false
		router.ServeHTTP(w, r)
		Log("should_pass:", calledTestHandler, " url:", mock.URL, " method:", mock.Method)
		if calledTestHandler != mock.CheckSuccess {
			t.Error("Handler check failed")
		}
	}

	paramMocks := []RequestMock{
		RequestMock{
			Method:       "GET",
			URL:          "/users/123/show",
			CheckSuccess: true,
			CheckParams:  true,
			CheckParamsValue: map[string][]string{
				"id": []string{"123"},
			},
		},
		RequestMock{
			Method:             "GET",
			URL:                "/test/route",
			CheckSuccess:       true,
			CheckCatchAll:      true,
			CheckCatchAllValue: "test/route",
		},
		RequestMock{
			Method:       "GET",
			URL:          "/users/123/show/something",
			CheckSuccess: true,
			CheckParams:  true,
			CheckParamsValue: map[string][]string{
				"id":   []string{"123"},
				"what": []string{"something"},
			},
		},
		RequestMock{
			Method:       "GET",
			URL:          "/users/123/show/something?test=55",
			CheckSuccess: true,
			CheckParams:  true,
			CheckParamsValue: map[string][]string{
				"id":   []string{"123"},
				"what": []string{"something"},
			},
		},
		RequestMock{
			Method:       "GET",
			URL:          "/users/123/show/something?what=55",
			CheckSuccess: true,
			CheckParams:  true,
			CheckParamsValue: map[string][]string{
				"id":   []string{"123"},
				"what": []string{"something"},
			},
		},
	}

	for _, mock := range paramMocks {
		r, _ := http.NewRequest(mock.Method, mock.URL, nil)
		w := httptest.NewRecorder()
		calledTestHandler = false
		finalParams = nil
		finalCatchAll = ""
		router.ServeHTTP(w, r)
		Log("should_pass:", calledTestHandler, " url:", mock.URL, " method:", mock.Method)
		if calledTestHandler != mock.CheckSuccess {
			t.Error("Handler check failed")
		} else {
			Log("params: ", finalParams)
			Log("catch-all:", finalCatchAll)
			if mock.CheckCatchAll == true && mock.CheckCatchAllValue != "" {
				if finalCatchAll != mock.CheckCatchAllValue {
					t.Error("Catch-all value check failed")
				}
			}
			if mock.CheckParams == true && mock.CheckParamsValue != nil {
				if !compareMaps(finalParams, mock.CheckParamsValue) {
					t.Error("Params value check failed")
				}
			}
		}
	}

}
