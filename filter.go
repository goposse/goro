// Goro
//
// Created by Yakka
// http://theyakka.com
//
// Copyright (c) 2018 Yakka LLC.
// All rights reserved.
// See the LICENSE file for licensing details and requirements.

package goro

import (
	"net/http"
)

// Filter is an interface that can be registered on the Router to apply custom
// logic to modify the Request or calling Context
type Filter interface {
	ExecuteFilter(req **http.Request)
}
