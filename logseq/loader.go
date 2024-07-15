package logseq

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gosimple/slug"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"export-logseq/graph"
)

const (
	branchBlockOpener    = "- "
	branchBlockContinuer = "  "
)

// Loader loads a Logseq graph from a directory.
type Loader struct {
	GraphDir string
	Graph    graph.Graph
}

func NewLoader(graphDir string) Loader {
	g := graph.NewGraph()
	g.GraphDir = graphDir

	return Loader{GraphDir: graphDir, Graph: g}
}

// LoadGraph loads a Logseq graph from a directory.
func LoadGraph(graphDir string) (graph.Graph, error) {
	log.Info("Loading Logseq graph from", graphDir)
	loader := NewLoader(graphDir)

	configFile := filepath.Join(graphDir, "logseq", "config.edn")
	if err := CheckConfig(configFile); err != nil {
		return loader.Graph, errors.Wrap(err, "loading Logseq config")
	}

	if err := loader.loadAssets(); err != nil {
		return loader.Graph, errors.Wrap(err, "loading assets")
	}

	pageDirs := []string{"pages", "journals"}

	for _, pageDir := range pageDirs {
		if err := loader.loadPagesFromDir(pageDir); err != nil {
			return loader.Graph, errors.Wrap(err, "loading pages from "+pageDir)
		}
	}

	loader.Graph.PutPagesInContext()

	return loader.Graph, nil
}

func (loader *Loader) LoadPage(pageFile string, graphPath string) (graph.Page, error) {
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
			return graph.Page{}, errors.New("decoding page name: " + decodeErr.Error())
		}

		fullPageName = escapedName
	}

	pathInGraph, err := filepath.Rel(graphPath, pageFile)
	if err != nil {
		return graph.Page{}, errors.Wrap(err, "calculating path in graph")
	}

	nameSteps := strings.Split(fullPageName, "/")
	title := nameSteps[len(nameSteps)-1]
	slugSteps := []string{}

	for _, step := range nameSteps {
		slugSteps = append(slugSteps, slug.Make(step))
	}

	pathInSite := strings.Join(slugSteps, "/")

	// Process each line of fullPageName
	file, err := os.Open(pageFile)
	if err != nil {
		return graph.Page{}, errors.New("opening page file: " + err.Error())
	}
	defer file.Close()

	lines, err := LoadPageLines(file)
	if err != nil {
		return graph.Page{}, errors.New("loading page lines: " + err.Error())
	}

	page := graph.Page{
		Name:        fullPageName,
		Title:       title,
		PathInGraph: pathInGraph,
		PathInSite:  pathInSite,
		Kind:        "page",
	}

	blocks, err := findBlocks(&page, lines)
	if err != nil {
		return graph.Page{}, errors.New("finding blocks: " + err.Error())
	}

	if len(blocks) == 0 {
		log.Warn("No root block found in page: ", fullPageName)

		blocks = []*graph.Block{graph.NewEmptyBlock()}
	}

	page.Root = blocks[0]
	page.AllBlocks = blocks

	return page, nil
}

func (loader *Loader) loadAssets() error {
	assetsDir := filepath.Join(loader.GraphDir, "assets")
	log.Info("Assets directory:", assetsDir)
	assetFiles, err := filepath.Glob(filepath.Join(assetsDir, "*.*"))

	if err != nil {
		return errors.Wrap(err, "listing asset files")
	}

	for _, assetFile := range assetFiles {
		relPath, err := filepath.Rel(assetsDir, assetFile)
		if err != nil {
			return errors.Wrap(err, "calculating relative path for asset")
		}

		asset := graph.NewAsset("/assets/" + relPath)
		err = loader.Graph.AddAsset(&asset)

		if err != nil {
			return errors.Wrap(err, "adding asset "+assetFile)
		}
	}

	return nil
}

func (loader *Loader) loadPagesFromDir(subdir string) error {
	g := &loader.Graph
	pagesDir := filepath.Join(g.GraphDir, subdir)
	log.Infof("Loading pages from %s", pagesDir)
	pageFiles, err := filepath.Glob(filepath.Join(pagesDir, "*.md"))

	if err != nil {
		return errors.Wrap(err, "listing page files")
	}

	for _, pageFile := range pageFiles {
		page, err := loader.LoadPage(pageFile, pagesDir)

		if err != nil {
			return errors.Wrap(err, "loading page "+pageFile)
		}

		err = g.AddPage(&page)
		if err != nil {
			return errors.Wrap(err, "adding loaded page "+pageFile)
		}
	}

	return nil
}

func findBlocks(page *graph.Page, lines []PageLine) ([]*graph.Block, error) {
	blocks := []*graph.Block{}
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
			block, err := graph.NewBlock(page, currentBlockLines, currentIndent)

			if err != nil {
				return nil, errors.Wrap(err, "creating new block")
			}

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
		block, err := graph.NewBlock(page, currentBlockLines, currentIndent)

		if err != nil {
			return nil, errors.Wrap(err, "creating block from remaining lines")
		}

		blocks = append(blocks, block)
		placeBlock(block, blockStack)
	}

	log.Debug("Blocks: ", blocks)

	return blocks, nil
}

func placeBlock(block *graph.Block, blockStack *BlockStack) *BlockStack {
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
