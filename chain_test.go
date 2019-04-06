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

func TestSimpleChain(t *testing.T) {
	path := "/chain/simple"
	Debug("Requesting", path, "...")
	execMockRequest(router, "GET", path)
	expectedSum := 13
	if sum != expectedSum {
		t.Error("Expected a sum of", expectedSum, "but got", sum)
	}
	resetState()
}

func TestMultiCallChain(t *testing.T) {
	path := "/chain/simple"
	Debug("Requesting", path, "...")
	execMockRequest(router, "GET", path)
	expectedSum := 13
	if sum != expectedSum {
		t.Error("Expected a sum of", expectedSum, "but got", sum)
	}
	sum = 0
	execMockRequest(router, "GET", path)
	if sum != expectedSum {
		t.Error("Expected a sum of", expectedSum, "but got", sum)
	}
	resetState()
}

func TestThenChain(t *testing.T) {
	path := "/chain/then"
	Debug("Requesting", path, "...")
	execMockRequest(router, "GET", path)
	expectedSum := 23
	if sum != expectedSum {
		t.Error("Expected a sum of", expectedSum, "but got", sum)
	}
	resetState()
}

func TestHaltChain(t *testing.T) {
	path := "/chain/halt"
	Debug("Requesting", path, "...")
	execMockRequest(router, "GET", path)
	expectedSum := 5
	if sum != expectedSum {
		t.Error("Expected a sum of", expectedSum, "but got", sum)
	}
	resetState()
}

func testThenHandler(_ *goro.HandlerContext) {
	Debug("Chain hit then")
	sum += 10
}

func chainHandler1(ch *goro.Chain, ctx *goro.HandlerContext) {
	Debug("Chain hit 1")
	sum = 1
	ch.Next(ctx)
}

func chainHandler2(ch *goro.Chain, ctx *goro.HandlerContext) {
	Debug("Chain hit 2")
	sum += 4
	ch.Next(ctx)
}

func chainHandler3(ch *goro.Chain, ctx *goro.HandlerContext) {
	Debug("Chain hit 3")
	sum += 8
	ch.Next(ctx)
}

func testHaltHandler(ch *goro.Chain, ctx *goro.HandlerContext) {
	Debug("Chain hit halt")
	ch.Halt(ctx)
}
