package logseq

// Linkable is an interface for types that can be linked to in a Logseq graph.
type Linkable interface {
	// PathInContext returns the path to the linkable in the context of the given graph.
	InContext(Graph) (string, error)
}
