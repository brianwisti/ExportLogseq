package logseq

import (
	"regexp"
	"strings"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type PropertyMap map[string]string
type Block struct {
	ID          string       `json:"id"`
	Content     string       `json:"content"`
	Properties  *PropertyMap `json:"properties,omitempty"`
	SourceLines []string     `json:"-"`
	Depth       int          `json:"-"`
	Position    int          `json:"-"`
	Callout     string       `json:"callout,omitempty"`
	Children    []*Block     `json:"children,omitempty"`
}

func (b *Block) ParseSourceLines() {
	propertyRe := regexp.MustCompile("^([a-zA-Z][a-zA-Z0-9_-]*):: (.*)")
	calloutOpenerRe := regexp.MustCompile("^#+BEGIN_([A-Z]+)")
	contentLines := []string{}
	properties := make(PropertyMap)
	var callout string
	inCallout := false

	for _, line := range b.SourceLines {
		propertyMatch := propertyRe.FindStringSubmatch(line)
		if propertyMatch != nil {
			prop_name, prop_value := propertyMatch[1], propertyMatch[2]
			properties[prop_name] = prop_value
			continue
		}

		if inCallout {
			if strings.HasPrefix(line, "#+END_"+callout) {
				inCallout = false
				continue
			}
		}

		calloutOpenerMatch := calloutOpenerRe.FindStringSubmatch(line)
		if calloutOpenerMatch != nil {
			callout = calloutOpenerMatch[1]
			inCallout = true
			continue
		}

		contentLines = append(contentLines, line)
	}

	if inCallout {
		log.Fatal("Unclosed callout")
	}

	uuidString, ok := properties["id"]
	if !ok {
		uuidString = uuid.New().String()
	}

	b.ID = uuidString
	b.Properties = &properties
	b.Callout = callout
	content := strings.Join(contentLines, "\n")
	b.Content = content

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