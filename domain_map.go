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
	"fmt"
	"net/http"
	"strings"
)

const DomainMapNakedSubdomainKey = ":naked"
const DomainMapWildcardSubdomainKey = ":wildcard"

// DomainMap - maps (sub)domains to routers
type DomainMap struct {
	domains     []string
	subdomains  []string
	routerMap   map[string]*Router
	hostMatches []string
	hasWildcard bool

	// NotFoundHandler - if the (sub)domain is not mapped, call this handler
	NotFoundHandler ContextHandler
}

// NewDomainMap - creates a new domain map for the provided domains
func NewDomainMap(domains ...string) *DomainMap {
	return &DomainMap{
		domains:     domains,
		subdomains:  []string{},
		routerMap:   map[string]*Router{},
		hasWildcard: false,
	}
}

// NewRouter - Creates a new Router, registers it in the domain map and returns it for use
func (dm *DomainMap) NewRouter(subdomainPattern string) *Router {
	router := NewRouter()
	subdomains := strings.Split(subdomainPattern, "|")
	for _, subdomain := range subdomains {
		key := subdomain
		if subdomain == "*" {
			key = DomainMapWildcardSubdomainKey
			dm.hasWildcard = true
		} else if subdomain == "<*>" {
			key = DomainMapNakedSubdomainKey
		}
		dm.AddRouter(key, router)
	}
	return router
}

// NewRouters - Creates a new Router for each of the defined subdomains and registers it
// with the DomainMap
func (dm *DomainMap) NewRouters(subdomains ...string) []*Router {
	var routers []*Router
	for _, subdomain := range subdomains {
		key := subdomain
		if subdomain == "*" {
			key = DomainMapWildcardSubdomainKey
			dm.hasWildcard = true
		} else if subdomain == "<*>" {
			key = DomainMapNakedSubdomainKey
		}
		router := NewRouter()
		dm.AddRouter(key, router)
		routers = append(routers, router)
	}
	return routers
}

// AddRouter - Register a router for a domain pattern (regex)
func (dm *DomainMap) AddRouter(subdomain string, router *Router) {
	dm.subdomains = append(dm.subdomains, subdomain)
	dm.routerMap[subdomain] = router
}

func (dm *DomainMap) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var router *Router
	subdomain, isMapped := dm.isMappedHost(req.Host)
	if isMapped {
		router = dm.routerMap[subdomain]
	} else if dm.isNakedHost(req.Host) {
		router = dm.routerMap[DomainMapNakedSubdomainKey]
	} else if dm.hasWildcard {
		router = dm.routerMap[DomainMapWildcardSubdomainKey]
	}
	if router == nil {
		if dm.NotFoundHandler != nil {
			dm.domainRouterNotFoundHandler().ServeHTTP(w, req)
		} else {
			// no handler & no matches found. fail.
			errorHandler(w, req, "Forbidden", http.StatusForbidden)
		}
		return
	}
	router.ServeHTTP(w, req)
}

func (dm DomainMap) isMappedHost(host string) (string, bool) {
	//for _, domainString := range dm.hostMatches {
	//	if domainString == host {
	//		return true
	//	}
	//}
	for _, domain := range dm.domains {
		for _, subdomain := range dm.subdomains {
			joinedDomain := fmt.Sprintf("%s.%s", subdomain, domain)
			if joinedDomain == host {
				//dm.hostMatches = append(dm.hostMatches, host)
				return subdomain, true
			}
		}
	}
	return "", false
}

func (dm DomainMap) isNakedHost(host string) bool {
	for _, registeredHost := range dm.domains {
		if host == registeredHost {
			return true
		}
	}
	return false
}

func (dm DomainMap) domainRouterNotFoundHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO - add additional context as to why it failed
		context := NewHandlerContext(r, w, nil)
		dm.NotFoundHandler(context)
	})
}
