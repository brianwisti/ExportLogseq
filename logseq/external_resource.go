package logseq

// ExternalResource represents a resource that is external to the Logseq graph.
type ExternalResource struct {
	Uri string `json:"uri"`
}

// InContext returns the URI of the external resource.
func (r ExternalResource) InContext(g Graph) (string, error) {
	return r.Uri, nil
}
