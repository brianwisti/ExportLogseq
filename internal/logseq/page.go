package logseq

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	log "github.com/sirupsen/logrus"
)

type PageLine struct {
	Content string
	Indent  int
}

type Page struct {
	Name        string   `json:"name"`
	PathInGraph string   `json:"-"`
	FullPath    string   `json:"-"`
	Blocks      []*Block `json:"blocks"`
}

func (p *Page) ParseBlocks() {
	for i := 0; i < len(p.Blocks); i++ {
		block := p.Blocks[i]
		block.ParseSourceLines()
	}
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

	page := Page{
		Name:        fullPageName,
		PathInGraph: pathInGraph,
		FullPath:    pageFile,
		Blocks:      blocks,
	}
	page.ParseBlocks()

	return page, nil
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

func findBlocks(lines []PageLine) ([]*Block, error) {
	branchBlockOpener := "- "
	branchBlockContinuer := "  "
	blocks := []*Block{}
	blockStack := &BlockStack{}
	currentBlockLines := []string{}
	currentIndent := 0

	for _, line := range lines {
		log.Debug("Line: ", line)
		// Skip empty block lines
		if line.Content == "-" {
			continue
		}

		if strings.HasPrefix(line.Content, branchBlockOpener) {
			// Remember the current block.
			block := Block{
				SourceLines: currentBlockLines,
				Depth:       currentIndent,
				Position:    len(blocks),
			}

			if block.Depth == 0 {
				blocks = append(blocks, &block)
			}
			blockStack = placeBlock(&block, blockStack)

			// Adjust for the root block not having a branch block marker.
			line.Indent = line.Indent + 1

			// Reset the current block and indent
			currentBlockLines = []string{}
			currentIndent = line.Indent
			line.Content = strings.TrimPrefix(line.Content, branchBlockOpener)
		} else if strings.HasPrefix(line.Content, branchBlockContinuer) {
			// Ensure that the current line is a continuation of a current block
			if len(currentBlockLines) == 0 {
				return blocks, errors.New("no block to continue: " + line.Content)
			}

			line.Content = strings.TrimPrefix(line.Content, branchBlockContinuer)
			// Adjust for the root block not having a branch block marker.
			line.Indent = line.Indent + 1
		}

		// Ensure that the current line is indented correctly
		if line.Indent != currentIndent {
			errMsg := fmt.Sprintf("mismatched indent: %v", line)
			return blocks, errors.New(errMsg)
		}

		currentBlockLines = append(currentBlockLines, line.Content)
	}

	// Remember the last block.
	if len(currentBlockLines) > 0 {
		block := Block{
			SourceLines: currentBlockLines,
			Depth:       currentIndent,
			Position:    len(blocks),
		}
		placeBlock(&block, blockStack)
	}
	log.Debug("Block stack: ", blockStack)
	log.Debug("Blocks: ", blocks)

	return blocks, nil
}

func placeBlock(block *Block, blockStack *BlockStack) *BlockStack {
	if block.Depth == 0 {
		blockStack.Push(block)
	} else {
		for topBlock := blockStack.Top(); topBlock != nil; topBlock = blockStack.Top() {
			if topBlock.Depth < block.Depth {
				topBlock.AddChild(block)
				log.Debug("Top block: ", topBlock)
				blockStack.Push(block)
				break
			}

			blockStack.Pop()
		}
	}
	return blockStack
}