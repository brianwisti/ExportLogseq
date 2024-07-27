package graph

import "path/filepath"

// Asset represents a non-note resource in a Logseq graph.
type Asset struct {
	Name string `json:"name"`
	Path string `json:"-"`
}

// NewAsset creates a new Asset with the given path in the graph.
func NewAsset(pathInGraph string) Asset {
	assetName := filepath.Base(pathInGraph)

	return Asset{
		Name: assetName,
		Path: pathInGraph,
	}
}
