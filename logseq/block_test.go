package logseq_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"export-logseq/logseq"
)

func TestParseSourceLines_WithProp(t *testing.T) {
	propName := "id"
	propValue := "123"
	propString := propName + ":: " + propValue

	block := &logseq.Block{
		SourceLines: []string{
			propString,
		},
	}
	block.ParseSourceLines()
	got, ok := block.Properties.Get("id")
	assert.True(t, ok)
	assert.Equal(t, got.Value, propValue)
}
