package logseq_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"export-logseq/graph"
	"export-logseq/logseq"
)

func TestBlockStack_NewBlockStack(t *testing.T) {
	bs := logseq.NewBlockStack()

	assert.NotNil(t, bs)
	assert.Empty(t, bs.Blocks)
}

func TestBlockStack_Push(t *testing.T) {
	bs := logseq.NewBlockStack()
	b := graph.NewEmptyBlock()

	bs.Push(b)

	assert.Contains(t, bs.Blocks, b)
}

func TestBlockStack_Pop(t *testing.T) {
	bs := logseq.NewBlockStack()
	b := graph.NewEmptyBlock()

	bs.Push(b)
	popped := bs.Pop()

	assert.Empty(t, bs.Blocks)
	assert.Equal(t, b, popped)
}

func TestBlockStack_Pop_Empty(t *testing.T) {
	bs := logseq.NewBlockStack()

	popped := bs.Pop()

	assert.Nil(t, popped)
}

func TestBlockStack_Top(t *testing.T) {
	bs := logseq.NewBlockStack()
	b := graph.NewEmptyBlock()

	bs.Push(b)
	top := bs.Top()

	assert.Equal(t, b, top)
	assert.Contains(t, bs.Blocks, b, "Top should not pop the block")
}

func TestBlockStack_Top_Empty(t *testing.T) {
	bs := logseq.NewBlockStack()

	top := bs.Top()

	assert.Nil(t, top)
}

func TestBlockStack_IsEmpty(t *testing.T) {
	bs := logseq.NewBlockStack()

	assert.True(t, bs.IsEmpty())
}

func TestBlockStack_IsEmpty_False(t *testing.T) {
	bs := logseq.NewBlockStack()
	b := graph.NewEmptyBlock()

	bs.Push(b)

	assert.False(t, bs.IsEmpty())
}
