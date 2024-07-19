package logseq_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"export-logseq/logseq"
)

func TestLoader_NewLoader(t *testing.T) {
	graphDir := "test_data/graph"
	l := logseq.NewLoader(graphDir)

	assert.NotNil(t, l)
	assert.NotNil(t, l.Graph)
}
