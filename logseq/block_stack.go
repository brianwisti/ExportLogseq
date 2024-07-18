package logseq

import (
	"export-logseq/graph"
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
