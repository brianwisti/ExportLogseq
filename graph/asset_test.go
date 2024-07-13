package graph_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"export-logseq/graph"
)

func AssetPath() string {
	return "assets/test.jpg"
}

func TestNewAsset(t *testing.T) {
	pathInGraph := AssetPath()
	asset := graph.NewAsset(pathInGraph)

	assert.Equal(t, pathInGraph, asset.PathInGraph)
}

func TestAsset_InContext(t *testing.T) {
	pathInGraph := AssetPath()
	asset := graph.Asset{
		PathInGraph: pathInGraph,
	}
	g := *graph.NewGraph()
	g.AddAsset(&asset)
	path, err := asset.InContext(g)

	assert.NoError(t, err)
	assert.Equal(t, pathInGraph, path)
}

func TestAsset_InContext_NotFound(t *testing.T) {
	asset := graph.Asset{
		PathInGraph: AssetPath(),
	}
	g := *graph.NewGraph()
	_, err := asset.InContext(g)

	assert.Error(t, err)
	assert.ErrorIs(t, err, graph.AssetNotFoundError{AssetPath: asset.PathInGraph})
}
