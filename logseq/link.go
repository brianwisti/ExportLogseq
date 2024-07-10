package logseq

type LinkType string

const (
	LinkTypeAsset    LinkType = "asset"
	LinkTypePage     LinkType = "page"
	LinkTypeResource LinkType = "resource"
	LinkTypeBlock    LinkType = "block"
)

// A Link connects two Linkable objects.
type Link struct {
	Raw       string   `json:"raw"`
	LinksFrom *Block   `json:"-"`
	LinkPath  string   `json:"link_path"`
	LinkType  LinkType `json:"link_type"`
	IsEmbed   bool     `json:"is_embed"`
	Label     string   `json:"label"`
}

// Convenience methods in case I change the implementation details.

// IsAsset returns true if the link is an asset link.
func (l *Link) IsAsset() bool {
	return l.LinkType == LinkTypeAsset
}

// IsBlock returns true if the link is a block link.
func (l *Link) IsBlock() bool {
	return l.LinkType == LinkTypeBlock
}

// IsPage returns true if the link is a page link.
func (l *Link) IsPage() bool {
	return l.LinkType == LinkTypePage
}

// IsResource returns true if the link is a resource link.
func (l *Link) IsResource() bool {
	return l.LinkType == LinkTypeResource
}
