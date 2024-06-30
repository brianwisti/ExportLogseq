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
	Name        string
	PathInGraph string
	FullPath    string
	Blocks      []Block
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

func findBlocks(lines []PageLine) ([]Block, error) {
	branchBlockOpener := "- "
	branchBlockContinuer := "  "
	blocks := []Block{}
	currentBlockLines := []string{}
	currentIndent := 0

	for _, line := range lines {
		// Skip empty block lines
		if line.Content == "-" {
			continue
		}

		if strings.HasPrefix(line.Content, branchBlockOpener) {
			// Adjust for the root block not having a branch block marker.
			line.Indent = line.Indent + 1

			// Remember the current block.
			if len(currentBlockLines) > 0 {
				block := Block{
					Content:  strings.Join(currentBlockLines, "\n"),
					Indent:   currentIndent,
					Position: len(blocks),
				}

				if block.Indent > 0 {
					// Find the parent block
					for i := block.Position - 1; i >= 0; i-- {
						parentBlock := &blocks[i]
						if parentBlock.Indent < block.Indent {
							block.Parent = parentBlock
							log.Debug("Parent found: ", parentBlock)
							parentBlock.Children = append(parentBlock.Children, &block)
							break
						}
					}
				}

				blocks = append(blocks, block)
			}

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
			Content:  strings.Join(currentBlockLines, "\n"),
			Indent:   currentIndent,
			Position: len(blocks),
		}

		if block.Indent > 0 {
			// Find the parent block
			for i := block.Position - 1; i >= 0; i-- {
				parentBlock := &blocks[i]
				if parentBlock.Indent < block.Indent {
					block.Parent = parentBlock
					break
				}
			}
		}

		blocks = append(blocks, block)
	}

	return blocks, nil
}
