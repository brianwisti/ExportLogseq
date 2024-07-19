package logseq

import (
	"export-logseq/graph"

	log "github.com/sirupsen/logrus"
)

// BlockStack helps track blocks being loaded from a page file.
type BlockStack struct {
	Blocks []*graph.Block
}

// NewBlockStack creates a new block stack.
func NewBlockStack() *BlockStack {
	return &BlockStack{
		Blocks: []*graph.Block{},
	}
}

// Top returns the block at the top of the stack, without popping it.
func (bs *BlockStack) Top() *graph.Block {
	if len(bs.Blocks) == 0 {
		return nil
	}

	return bs.Blocks[len(bs.Blocks)-1]
}

// IsEmpty returns true if the stack is empty.
func (bs *BlockStack) IsEmpty() bool {
	return len(bs.Blocks) == 0
}

// Push a block onto the stack.
func (bs *BlockStack) Push(b *graph.Block) {
	bs.Blocks = append(bs.Blocks, b)
}

// Pop a block from the stack.
func (bs *BlockStack) Pop() *graph.Block {
	if len(bs.Blocks) == 0 {
		return nil
	}

	lastIndex := len(bs.Blocks) - 1
	top := bs.Blocks[lastIndex]
	bs.Blocks = bs.Blocks[:lastIndex]

	return top
}

func PlaceBlock(block *graph.Block, blockStack *BlockStack) *BlockStack {
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
