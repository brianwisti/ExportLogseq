package logseq_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"export-logseq/logseq"
)

func TestBlockContent_IsCodeBlock(t *testing.T) {
	codeBlockMarkdown := "```\ncode\n```"
	blockContent := logseq.BlockContent{
		Markdown: codeBlockMarkdown,
	}
	assert.True(t, blockContent.IsCodeBlock())
}
