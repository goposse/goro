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
	"testing"
)

func TestSimpleRoute(t *testing.T) {
	expectHitResult(t, router, "GET", "/")
}

func TestMissedRoute(t *testing.T) {
	expectNotHitResult(t, router, "GET", "/turnip")
}

func TestNamedParamRoute(t *testing.T) {
	expectHitResult(t, router, "GET", "/users/A139871")
}

func TestMultiNamedParamRoute(t *testing.T) {
	expectHitResult(t, router, "GET", "/users/A139871/action/call")
}

func TestVariableSubRoute(t *testing.T) {
	expectHitResult(t, router, "GET", "/colors/blue")
}

func TestVariableSubMissRoute(t *testing.T) {
	expectNotHitResult(t, router, "GET", "/colors/red")
}

func testHandler(_ *goro.HandlerContext) {
	wasHit = true
}

func testParamsHandler(ctx *goro.HandlerContext) {
	wasHit = true
}
