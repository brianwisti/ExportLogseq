package graph

import "path/filepath"

// Asset represents a non-note resource in a Logseq graph.
type Asset struct {
	Name        string `json:"name"`
	PathInGraph string `json:"path_in_graph"`
	PathInSite  string `json:"path_in_site"`
}

// NewAsset creates a new Asset with the given path in the graph.
func NewAsset(pathInGraph string) Asset {
	assetName := filepath.Base(pathInGraph)

	return Asset{
		PathInGraph: pathInGraph,
		PathInSite:  pathInGraph,
		Name:        assetName,
	}
}

// InContext returns the path to the asset in the context of the given graph.
func (a Asset) InContext(g Graph) (string, error) {
	knownAsset, ok := g.FindAsset(a.PathInGraph)
	if !ok {
		return "", AssetNotFoundError{a.PathInGraph}
	}

	return knownAsset.PathInGraph, nil
}
