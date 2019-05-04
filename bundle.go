package goro

type Bundle interface {
	PreFilters() []*Filter
	PostFilters() []*Filter
	Routes() []*Route
}
