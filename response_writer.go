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
	"net/http"
)

type CheckedResponseWriter struct {
	http.ResponseWriter
	headerWritten bool
}

func NewCheckedResponseWriter(w http.ResponseWriter) *CheckedResponseWriter {
	return &CheckedResponseWriter{ResponseWriter: w}
}

func (w *CheckedResponseWriter) WriteHeader(status int) {
	if w.headerWritten {
		return
	}
	w.ResponseWriter.WriteHeader(status)
	w.headerWritten = true
}

func (w *CheckedResponseWriter) Write(b []byte) (int, error) {
	if !w.headerWritten {
		w.WriteHeader(http.StatusOK)
		w.headerWritten = true
	}
	return w.ResponseWriter.Write(b)
}
