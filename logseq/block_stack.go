package logseq

import (
	"export-logseq/graph"
)

// BlockStack helps track blocks being loaded from a page file.
type BlockStack struct {
	Blocks []*graph.Block
}

func (bs *BlockStack) Push(b *graph.Block) {
	bs.Blocks = append(bs.Blocks, b)
}

func (bs *BlockStack) Pop() *graph.Block {
	if len(bs.Blocks) == 0 {
		return nil
	}

	lastIndex := len(bs.Blocks) - 1
	top := bs.Blocks[lastIndex]
	bs.Blocks = bs.Blocks[:lastIndex]

	return top
}

func (bs *BlockStack) Top() *graph.Block {
	if len(bs.Blocks) == 0 {
		return nil
	}

	return bs.Blocks[len(bs.Blocks)-1]
}

func (bs *BlockStack) IsEmpty() bool {
	return len(bs.Blocks) == 0
}
