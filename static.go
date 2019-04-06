// Goro
//
// Created by Yakka
// http://theyakka.com
//
// Copyright (c) 2019 Yakka LLC.
// All rights reserved.
// See the LICENSE file for licensing details and requirements.

package goro

// StaticLocation is a holder for static location information
type StaticLocation struct {
	// root is the root (source) location
	root string

	// prefix is a path prefix to applied when matching
	prefix string
}
