package logseq

import (
	"bufio"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/pkg/errors"
)

type PageLine struct {
	Content string
	Indent  int
}

func LoadPageLines(file *os.File) ([]PageLine, error) {
	var lines []PageLine

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		fullLength := utf8.RuneCountInString(line)
		lineContent := strings.TrimLeft(line, "\t")
		indent := fullLength - utf8.RuneCountInString(lineContent)
		lines = append(lines, PageLine{
			Content: lineContent,
			Indent:  indent,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, errors.Wrap(err, "scanning page lines")
	}

	return lines, nil
}
