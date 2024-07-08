package logseq_test

import (
	"testing"

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
