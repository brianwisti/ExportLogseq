package logseq_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"export-logseq/logseq"
)

func TestPageLine_NewPageLine(t *testing.T) {
	pageLineTests := []struct {
		Text string
		Want logseq.PageLine
	}{
		{"", logseq.PageLine{}},
		{"text", logseq.PageLine{Content: "text", Indent: 0}},
		{"  continuation", logseq.PageLine{Content: "  continuation", Indent: 0}},
		{"\t- block", logseq.PageLine{Content: "- block", Indent: 1}},
		{"\t  continuation", logseq.PageLine{Content: "  continuation", Indent: 1}},
	}

	for _, tt := range pageLineTests {
		pl := logseq.NewPageLine(tt.Text)

		assert.Equal(t, tt.Want, pl)
	}
}

func TestPageLine_LoadPageLines(t *testing.T) {
	pageLines := []string{
		"line 1",
		"  line 2",
		"\tline 3",
		"\t  line 4",
		"line 5",
	}

	expected := []logseq.PageLine{
		{Content: "line 1", Indent: 0},
		{Content: "  line 2", Indent: 0},
		{Content: "line 3", Indent: 1},
		{Content: "  line 4", Indent: 1},
		{Content: "line 5", Indent: 0},
	}

	var buffer bytes.Buffer

	for _, line := range pageLines {
		buffer.WriteString(line)
		buffer.WriteString("\n")
	}

	reader := bytes.NewReader(buffer.Bytes())
	pl, err := logseq.LoadPageLines(reader)

	assert.NoError(t, err)
	assert.Equal(t, expected, pl)
}
