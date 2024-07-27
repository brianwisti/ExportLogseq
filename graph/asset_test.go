package graph_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"export-logseq/graph"
)

func AssetPath() string {
	return "assets/test.jpg"
}

func TestNewAsset(t *testing.T) {
	pathInGraph := AssetPath()
	expectedName := filepath.Base(pathInGraph)
	asset := graph.NewAsset(pathInGraph)

	assert.Equal(t, pathInGraph, asset.Path)
	assert.Equal(t, expectedName, asset.Name)
}
