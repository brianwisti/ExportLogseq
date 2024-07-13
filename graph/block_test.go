package graph_test

import (
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"export-logseq/graph"
)

func TestParseSourceLines_WithProp(t *testing.T) {
	page := graph.NewEmptyPage()
	propName := "id"
	propValue := "123"
	propString := propName + ":: " + propValue
	block, err := graph.NewBlock(&page, []string{propString}, 0)

	require.NoError(t, err)

	got, ok := block.Properties.Get("id")

	assert.True(t, ok)
	assert.Equal(t, got.Value, propValue)
}

func TestBlock_AddChild(t *testing.T) {
	block := graph.NewEmptyBlock()
	child := graph.NewEmptyBlock()
	block.AddChild(child)

	assert.Contains(t, block.Children, child)
	assert.Equal(t, block, child.Parent)
}

func TestBlock_IsPublic_Default(t *testing.T) {
	block := graph.NewEmptyBlock()

	assert.False(t, block.IsPublic())
}

func TestBlock_IsPublic_FromProp(t *testing.T) {
	isPublicTests := []struct {
		PropValue string
		want      bool
	}{
		{"true", true},
		{"false", false},
		{"", false},
	}

	block := graph.NewEmptyBlock()
	for _, tt := range isPublicTests {
		block.Properties.Set("public", tt.PropValue)

		assert.Equal(t, tt.want, block.IsPublic())
	}

	assert.False(t, block.IsPublic())
}

func TestBlock_IsPublic_Cascading(t *testing.T) {
	isPublicTests := []struct {
		PropValue string
		want      bool
	}{
		{"true", true},
		{"false", false},
		{"", false},
	}

	for _, tt := range isPublicTests {
		block := graph.NewEmptyBlock()
		child := graph.NewEmptyBlock()

		block.Properties.Set("public", tt.PropValue)
		block.AddChild(child)

		assert.Equal(t, tt.want, child.IsPublic())
	}
}

func TestBlock_IsPublic_OverridesParent(t *testing.T) {
	block := graph.NewEmptyBlock()
	child := graph.NewEmptyBlock()

	block.Properties.Set("public", "false")
	block.AddChild(child)
	child.Properties.Set("public", "true")

	assert.True(t, child.IsPublic())
}

func TestBlock_Links(t *testing.T) {
	block := graph.NewEmptyBlock()
	link := graph.Link{
		LinkPath: gofakeit.URL(),
		Label:    gofakeit.Phrase(),
		LinkType: graph.LinkTypeResource,
		IsEmbed:  false,
	}
	link, _ = block.Content.AddLink(link)
	links := block.Links()

	assert.Contains(t, links, link)
}

func TestBlock_Links_Empty(t *testing.T) {
	block := graph.NewEmptyBlock()
	links := block.Links()

	assert.NotNil(t, links)
	assert.Empty(t, links)
}
