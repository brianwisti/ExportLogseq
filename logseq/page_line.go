package logseq

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/pkg/errors"
)

type PageLine struct {
	Content string
	Indent  int
}

// NewPageLine creates a new PageLine from a string.
func NewPageLine(line string) PageLine {
	fullLength := utf8.RuneCountInString(line)
	lineContent := strings.TrimLeft(line, "\t")
	indent := fullLength - utf8.RuneCountInString(lineContent)

	return PageLine{
		Content: lineContent,
		Indent:  indent,
	}
}

func LoadPageLines(r io.Reader) ([]PageLine, error) {
	var lines []PageLine

	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		pageLine := NewPageLine(scanner.Text())
		lines = append(lines, pageLine)
	}

	if err := scanner.Err(); err != nil {
		return nil, errors.Wrap(err, "scanning page lines")
	}

	return lines, nil
}

func (pl PageLine) String() string {
	return fmt.Sprintf("<PageLine: Depth=%d; Content=%s>", pl.Indent, pl.Content)
}
