package goro_test

import (
	"github.com/theyakka/goro"
	"testing"
)

func TestSubdomains(t *testing.T) {
	domains := goro.NewDomainMap("localhost.local")
	// <*> = naked domain
	router := domains.NewRouter("www|<*>")
	router.GET("/hello").Handle(helloHandler)
	router2 := domains.NewRouter("funky|chicken")
	router2.GET("/colors").Handle(testHandler)
	router3 := domains.NewRouter("*")
	router3.GET("/wildcard").Handle(testHandler)
	expectHitResult(t, domains, "GET", "http://catchall.localhost.local/wildcard")
	expectHitResult(t, domains, "GET", "http://www.localhost.local/hello")
	expectNotHitResult(t, domains, "GET", "http://chicken.localhost.local/hello")
	expectHitResult(t, domains, "GET", "http://chicken.localhost.local/colors")
	expectHitResult(t, domains, "GET", "http://chicken.localhost.local/colors")
	expectHitResult(t, domains, "GET", "http://localhost.local/hello")
	expectNotHitResult(t, domains, "GET", "http://chicken.localhost.local/wildcard")
	expectNotHitResult(t, domains, "GET", "http://localhost.local/wildcard")
}

func helloHandler(_ *goro.HandlerContext) {
	wasHit = true
}
