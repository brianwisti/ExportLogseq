package graph_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"export-logseq/graph"
)

func TestLink_IsBlock(t *testing.T) {
	link := graph.Link{
		LinkPath: "test-block",
		LinkType: graph.LinkTypeBlock,
	}

	assert.True(t, link.IsBlock())

	link.LinkType = graph.LinkTypePage
	assert.False(t, link.IsBlock())
}

func TestLink_IsAsset(t *testing.T) {
	link := graph.Link{
		LinkPath: "test-asset",
		LinkType: graph.LinkTypeAsset,
	}

	assert.True(t, link.IsAsset())

	link.LinkType = graph.LinkTypePage
	assert.False(t, link.IsAsset())
}

func TestLink_IsPage(t *testing.T) {
	link := graph.Link{
		LinkPath: "test-page",
		LinkType: graph.LinkTypePage,
	}

	assert.True(t, link.IsPage())

	link.LinkType = graph.LinkTypeAsset
	assert.False(t, link.IsPage())
}

func TestLink_IsResource(t *testing.T) {
	link := graph.Link{
		LinkPath: "test-resource",
		LinkType: graph.LinkTypeResource,
	}

	assert.True(t, link.IsResource())

	link.LinkType = graph.LinkTypeAsset
	assert.False(t, link.IsResource())
}
