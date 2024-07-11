package logseq_test

import (
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"

	"export-logseq/logseq"
)

func TestParseSourceLines_WithProp(t *testing.T) {
	page := logseq.NewEmptyPage()
	propName := "id"
	propValue := "123"
	propString := propName + ":: " + propValue
	block := logseq.NewBlock(&page, []string{propString}, 0)
	got, ok := block.Properties.Get("id")

	assert.True(t, ok)
	assert.Equal(t, got.Value, propValue)
}

func TestBlock_AddChild(t *testing.T) {
	block := logseq.NewEmptyBlock()
	child := logseq.NewEmptyBlock()
	block.AddChild(child)

	assert.Contains(t, block.Children, child)
	assert.Equal(t, block, child.Parent)
}

func TestBlock_IsPublic_Default(t *testing.T) {
	block := logseq.NewEmptyBlock()

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

	block := logseq.NewEmptyBlock()
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
		block := logseq.NewEmptyBlock()
		block.Properties.Set("public", tt.PropValue)
		child := logseq.NewEmptyBlock()
		block.AddChild(child)

		assert.Equal(t, tt.want, child.IsPublic())
	}
}

func TestBlock_IsPublic_OverridesParent(t *testing.T) {
	block := logseq.NewEmptyBlock()
	block.Properties.Set("public", "false")
	child := logseq.NewEmptyBlock()
	block.AddChild(child)
	child.Properties.Set("public", "true")

	assert.True(t, child.IsPublic())
}

func TestBlock_Links(t *testing.T) {
	block := logseq.NewEmptyBlock()
	link := logseq.Link{
		LinkPath: gofakeit.URL(),
		Label:    gofakeit.Phrase(),
		LinkType: logseq.LinkTypeResource,
		IsEmbed:  false,
	}
	link, _ = block.Content.AddLink(link)
	links := block.Links()

	assert.Contains(t, links, link)
}

func TestBlock_Links_Empty(t *testing.T) {
	block := logseq.NewEmptyBlock()
	links := block.Links()

	assert.NotNil(t, links)
	assert.Empty(t, links)
}
