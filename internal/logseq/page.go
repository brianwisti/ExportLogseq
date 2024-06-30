package logseq

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	log "github.com/sirupsen/logrus"
)

type Page struct {
	Name        string
	PathInGraph string
	FullPath    string
}

type PageLine struct {
	Content string
	Indent  int
}

type Block struct {
	ContentLines []string
	Indent       int
}

func (b Block) String() string {
	return strings.Join(b.ContentLines, "\n")
}

func LoadPage(pageFile string, graphPath string) (Page, error) {
	baseName := filepath.Base(pageFile)
	fullPageFileName := strings.ReplaceAll(baseName, "___", "/")
	fullPageName := strings.TrimSuffix(fullPageFileName, ".md")

	pathInGraph, err := filepath.Rel(graphPath, pageFile)
	if err != nil {
		return Page{}, errors.New("calculating path in graph: " + err.Error())
	}

	// Process each line of fullPageName
	file, err := os.Open(pageFile)
	if err != nil {
		return Page{}, errors.New("opening page file: " + err.Error())
	}
	defer file.Close()

	lines, err := loadPageLines(file)
	if err != nil {
		return Page{}, errors.New("loading page lines: " + err.Error())
	}

	branchBlockOpener := "- "
	branchBlockContinuer := "  "
	blocks := []Block{}
	currentBlockLines := []string{}
	currentIndent := 0

	for _, line := range lines {
		if strings.HasPrefix(line.Content, branchBlockOpener) {
			// Remember and reset the current block.
			if len(currentBlockLines) > 0 {
				blocks = append(blocks, Block{
					ContentLines: currentBlockLines,
					Indent:       currentIndent,
				})
			}
			currentBlockLines = []string{}
			currentIndent = line.Indent
			line.Content = strings.TrimPrefix(line.Content, branchBlockOpener)
		} else if strings.HasPrefix(line.Content, branchBlockContinuer) {
			// Ensure that the current line is a continuation of the current block
			// by checking that the current block has at least one line and the current
			// line has the same indent as the block.
			if len(currentBlockLines) == 0 {
				return Page{}, errors.New("no block to continue: " + line.Content)
			}

			line.Content = strings.TrimPrefix(line.Content, branchBlockContinuer)
		}

		if line.Indent != currentIndent {
			return Page{}, errors.New("mismatched indent: " + line.Content)
		}

		currentBlockLines = append(currentBlockLines, line.Content)
	}

	// Remember the last block.
	if len(currentBlockLines) > 0 {
		blocks = append(blocks, Block{
			ContentLines: currentBlockLines,
			Indent:       currentIndent,
		})
	}

	for _, block := range blocks {
		log.Info(block)
	}

	return Page{
		Name:        fullPageName,
		PathInGraph: pathInGraph,
		FullPath:    pageFile,
	}, nil
}

func loadPageLines(file *os.File) ([]PageLine, error) {
	scanner := bufio.NewScanner(file)
	var lines []PageLine

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
		return nil, errors.New("scanning page lines: " + err.Error())
	}

	return lines, nil
}
