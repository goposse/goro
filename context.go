package goro

import "net/http"

type HandlerContext struct {
	Request        *http.Request
	ResponseWriter http.ResponseWriter
	Parameters     *Parameters
	Meta           map[string]interface{}
	Path           string
	CatchAllValue  string
	Errors         []ErrorMap
	router         *Router
}
