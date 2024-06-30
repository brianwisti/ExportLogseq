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
	Blocks      []Block
}

type PageLine struct {
	Content string
	Indent  int
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

	blocks, err := findBlocks(lines)
	if err != nil {
		return Page{}, errors.New("finding blocks: " + err.Error())
	}

	for _, block := range blocks {
		log.Info(block)
	}

	return Page{
		Name:        fullPageName,
		PathInGraph: pathInGraph,
		FullPath:    pageFile,
		Blocks:      blocks,
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
