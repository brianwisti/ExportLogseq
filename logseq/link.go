package logseq

// Linkable is an interface for types that can be linked to in a Logseq graph.
type Linkable interface {
	// PathInContext returns the path to the linkable in the context of the given graph.
	InContext(Graph) (string, error)
}

// ExternalResource represents a resource that is external to the Logseq graph.
type ExternalResource struct {
	Uri string `json:"uri"`
}

// InContext returns the URI of the external resource.
func (r ExternalResource) InContext(g Graph) (string, error) {
	return r.Uri, nil
}

// A Link connects two Linkable objects.
type Link struct {
	Raw       string   `json:"-"`
	LinksFrom Linkable `json:"from"`
	LinksTo   Linkable `json:"to"`
	IsEmbed   bool     `json:"is_embed"`
	Label     string   `json:"label"`
}
