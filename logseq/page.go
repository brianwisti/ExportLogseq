package logseq

import (
	"bufio"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/gosimple/slug"
	log "github.com/sirupsen/logrus"
)

type PageLine struct {
	Content string
	Indent  int
}

type Page struct {
	Name        string   `json:"name"`
	PathInSite  string   `json:"path"`
	PathInGraph string   `json:"-"`
	Kind        string   `json:"kind"`
	Root        *Block   `json:"root"`
	AllBlocks   []*Block `json:"-"`
}

func NewEmptyPage() Page {
	root := NewEmptyBlock()

	return Page{
		Kind:      "page",
		Root:      root,
		AllBlocks: []*Block{root},
	}
}

func LoadPage(pageFile string, graphPath string) (Page, error) {
	baseName := filepath.Base(pageFile)
	fullPageFileName := strings.ReplaceAll(baseName, "___", "/")
	fullPageName := strings.TrimSuffix(fullPageFileName, ".md")
	journalDateRe := regexp.MustCompile(`^\d{4}_\d{2}_\d{2}$`)
	titleIsJournalDate := journalDateRe.MatchString(fullPageName)

	if titleIsJournalDate {
		fullPageName = strings.Replace(fullPageName, "_", "-", -1)
	} else {
		escapedName, decodeErr := url.QueryUnescape(fullPageName)
		if decodeErr != nil {
			return Page{}, errors.New("decoding page name: " + decodeErr.Error())
		}

		fullPageName = escapedName
	}

	pathInGraph, err := filepath.Rel(graphPath, pageFile)
	if err != nil {
		return Page{}, errors.New("calculating path in graph: " + err.Error())
	}

	nameSteps := strings.Split(fullPageName, "/")
	slugSteps := []string{}
	for _, step := range nameSteps {
		slugSteps = append(slugSteps, slug.Make(step))
	}
	pathInSite := strings.Join(slugSteps, "/")

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

	page := Page{
		Name:        fullPageName,
		PathInGraph: pathInGraph,
		PathInSite:  pathInSite,
		Kind:        "page",
	}

	blocks, err := findBlocks(&page, lines)
	if err != nil {
		return Page{}, errors.New("finding blocks: " + err.Error())
	}
	if len(blocks) == 0 {
		log.Warn("No root block found in page: ", fullPageName)
		blocks = []*Block{NewEmptyBlock()}
	}

	page.Root = blocks[0]
	page.AllBlocks = blocks

	return page, nil
}

func (p *Page) InContext(Graph) (string, error) {
	return "/" + p.PathInSite, nil
}

func (p *Page) Properties() *PropertyMap {
	return p.Root.Properties
}

// SetRoot assign's page root block and sets AllBlocks to root's branches
func (p *Page) SetRoot(root *Block) {
	p.Root = root
	p.AllBlocks = []*Block{}
	p.addTree(root)
}

func (p *Page) addTree(block *Block) {
	p.AllBlocks = append(p.AllBlocks, block)
	for _, child := range block.Children {
		p.addTree(child)
	}
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

func findBlocks(page *Page, lines []PageLine) ([]*Block, error) {
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
			block := NewBlock(page, currentBlockLines, currentIndent)
			blocks = append(blocks, block)
			blockStack = placeBlock(block, blockStack)

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
		block := NewBlock(page, currentBlockLines, currentIndent)
		blocks = append(blocks, block)
		placeBlock(block, blockStack)
	}
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
