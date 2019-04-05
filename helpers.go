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
	"strings"
)

// CleanPath - returns a path value with the following modifications:
//	1. replaces any '\' with '/'
//	2. replaces any '//' with '/'
//	3. adds a leading '/' (if missing)
func CleanPath(path string) string {
	cleanPath := path
	// replace any non-unix path separators
	cleanPath = strings.Replace(cleanPath, "\\", "/", -1)
	// replace double separators
	cleanPath = strings.Replace(cleanPath, "//", "/", -1)
	// add leading slash if missing
	if !strings.HasPrefix(cleanPath, "/") {
		cleanPath = "/" + cleanPath
	}
	return cleanPath
}
