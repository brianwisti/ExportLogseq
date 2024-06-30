package logseq

import (
	"errors"
	"fmt"
	"strings"

	log "github.com/charmbracelet/log"
)

type Block struct {
	ContentLines []string
	Indent       int
	Position     int
	Parent       *Block
	Children     []*Block
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
					ContentLines: currentBlockLines,
					Indent:       currentIndent,
					Position:     len(blocks),
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
			ContentLines: currentBlockLines,
			Indent:       currentIndent,
			Position:     len(blocks),
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
