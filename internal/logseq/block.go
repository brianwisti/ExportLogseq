package logseq

type Block struct {
	Content  string
	Indent   int
	Position int
	Parent   *Block
	Children []*Block
}
