package logseq

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

type Block struct {
	ID         string        `json:"id"`
	Page       *Page         `json:"-"` // Page that contains this block
	Content    *BlockContent `json:"content"`
	Properties *PropertyMap  `json:"properties,omitempty"`
	Depth      int           `json:"-"`
	Parent     *Block        `json:"-"`
	Children   []*Block      `json:"children,omitempty"`
}

func NewEmptyBlock() *Block {
	content := NewEmptyBlockContent()
	block := Block{
		ID:         uuid.New().String(),
		Content:    content,
		Properties: NewPropertyMap(),
	}
	content.Block = &block

	return &block
}

func NewBlock(page *Page, sourceLines []string, depth int) *Block {
	propertyRe := regexp.MustCompile("^([a-zA-Z][a-zA-Z0-9_-]*):: (.*)")
	contentLines := []string{}
	properties := NewPropertyMap()

	for _, line := range sourceLines {
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

	block := Block{
		ID:         uuidString,
		Page:       page,
		Depth:      depth,
		Properties: properties,
	}
	content := strings.Join(contentLines, "\n")
	blockContent := NewBlockContent(&block, content)
	block.Content = blockContent

	return &block
}

func (b *Block) AddChild(child *Block) {
	b.Children = append(b.Children, child)
	child.Parent = b
}

func (b *Block) InContext(g Graph) (string, error) {
	pagePath, err := b.Page.InContext(g)
	if err != nil {
		return "Setting block context", err
	}

	return pagePath + "#" + b.ID, nil
}

func (b *Block) IsPublic() bool {
	if publicProp, ok := b.Properties.Get("public"); ok {
		return publicProp.Bool()
	}

	if b.Parent != nil {
		return b.Parent.IsPublic()
	}

	return false
}

// PageLinks returns all page links found in the block
func (b *Block) PageLinks() []*Link {
	return b.Content.PageLinks
}

// ResourceLinks returns all resource links found in the block
func (b *Block) ResourceLinks() []*Link {
	return b.Content.ResourceLinks
}

func (b *Block) String() string {
	pageName := "NO PAGE"

	if b.Page != nil {
		pageName = b.Page.Name
	}

	return fmt.Sprintf("<Block: %s#%s>", pageName, b.ID)
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
