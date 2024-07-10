package logseq_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"export-logseq/logseq"
)

func TestLink_IsBlock(t *testing.T) {
	link := logseq.Link{
		LinkPath: "test-block",
		LinkType: logseq.LinkTypeBlock,
	}

	assert.True(t, link.IsBlock())

	link.LinkType = logseq.LinkTypePage
	assert.False(t, link.IsBlock())
}

func TestLink_IsAsset(t *testing.T) {
	link := logseq.Link{
		LinkPath: "test-asset",
		LinkType: logseq.LinkTypeAsset,
	}

	assert.True(t, link.IsAsset())

	link.LinkType = logseq.LinkTypePage
	assert.False(t, link.IsAsset())
}

func TestLink_IsPage(t *testing.T) {
	link := logseq.Link{
		LinkPath: "test-page",
		LinkType: logseq.LinkTypePage,
	}

	assert.True(t, link.IsPage())

	link.LinkType = logseq.LinkTypeAsset
	assert.False(t, link.IsPage())
}

func TestLink_IsResource(t *testing.T) {
	link := logseq.Link{
		LinkPath: "test-resource",
		LinkType: logseq.LinkTypeResource,
	}

	assert.True(t, link.IsResource())

	link.LinkType = logseq.LinkTypeAsset
	assert.False(t, link.IsResource())
}
