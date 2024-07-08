package logseq

import (
	"regexp"
	"strings"

	"github.com/google/uuid"
)

type Block struct {
	ID          string        `json:"id"`
	Content     *BlockContent `json:"content"`
	Properties  *PropertyMap  `json:"properties,omitempty"`
	SourceLines []string      `json:"-"`
	Depth       int           `json:"-"`
	Position    int           `json:"-"`
	Children    []*Block      `json:"children,omitempty"`
}

func NewEmptyBlock() *Block {
	return &Block{
		Content:    EmptyBlockContent(),
		Properties: NewPropertyMap(),
	}
}

func (b *Block) ParseSourceLines() {
	propertyRe := regexp.MustCompile("^([a-zA-Z][a-zA-Z0-9_-]*):: (.*)")
	contentLines := []string{}
	properties := NewPropertyMap()

	for _, line := range b.SourceLines {
		propertyMatch := propertyRe.FindStringSubmatch(line)
		if propertyMatch != nil {
			prop_name, prop_value := propertyMatch[1], propertyMatch[2]
			properties.Set(prop_name, prop_value)
			continue
		}

		contentLines = append(contentLines, line)
	}

	uuidString := uuid.New().String()
	idProp, _ := properties.Get("id")

	if idProp != nil {
		uuidString = idProp.Value
	} else {
		properties.Set("id", uuidString)
	}

	b.ID = uuidString
	b.Properties = properties
	content := strings.Join(contentLines, "\n")
	blockContent := BlockContentFromRawSource(content)
	b.Content = blockContent

	// parse children
	for _, child := range b.Children {
		child.ParseSourceLines()
	}
}

func (b *Block) AddChild(child *Block) {
	b.Children = append(b.Children, child)
}

type BlockStack struct {
	Blocks []*Block
}

func (bs *BlockStack) Push(b *Block) {
	bs.Blocks = append(bs.Blocks, b)
}

func (bs *BlockStack) Pop() *Block {
	if len(bs.Blocks) == 0 {
		return nil
	}

	lastIndex := len(bs.Blocks) - 1
	top := bs.Blocks[lastIndex]
	bs.Blocks = bs.Blocks[:lastIndex]
	return top
}

func (bs *BlockStack) Top() *Block {
	if len(bs.Blocks) == 0 {
		return nil
	}

	return bs.Blocks[len(bs.Blocks)-1]
}

func (bs *BlockStack) IsEmpty() bool {
	return len(bs.Blocks) == 0
}
