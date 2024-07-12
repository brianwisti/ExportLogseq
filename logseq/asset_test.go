package logseq_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"export-logseq/logseq"
)

func AssetPath() string {
	return "assets/test.jpg"
}

func TestNewAsset(t *testing.T) {
	pathInGraph := AssetPath()
	asset := logseq.NewAsset(pathInGraph)

	assert.Equal(t, pathInGraph, asset.PathInGraph)
}

func TestAsset_InContext(t *testing.T) {
	pathInGraph := AssetPath()
	asset := logseq.Asset{
		PathInGraph: pathInGraph,
	}
	g := *logseq.NewGraph()
	g.AddAsset(&asset)
	path, err := asset.InContext(g)

	assert.NoError(t, err)
	assert.Equal(t, pathInGraph, path)
}

func TestAsset_InContext_NotFound(t *testing.T) {
	asset := logseq.Asset{
		PathInGraph: AssetPath(),
	}
	g := *logseq.NewGraph()
	_, err := asset.InContext(g)

	assert.Error(t, err)
	assert.ErrorIs(t, err, logseq.AssetNotFoundError{AssetPath: asset.PathInGraph})
}
