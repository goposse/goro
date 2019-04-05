package goro

// StaticLocation is a holder for static location information
type StaticLocation struct {
	// root is the root (source) location
	root string

	// prefix is a path prefix to applied when matching
	prefix string
}
