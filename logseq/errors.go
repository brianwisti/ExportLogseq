package logseq

type DisconnectedPageError struct {
	PageName string
}

func (p DisconnectedPageError) Error() string {
	return "page not found in graph: " + p.PageName
}

// ErrorDuplicatePageLink is returned when a block content already has a link to a page and another link is added.
type ErrorDuplicatePageLink struct {
	PageName string
}

func (e ErrorDuplicatePageLink) Error() string {
	return "duplicate page link: " + e.PageName
}

// ErrorDuplicateResourceLink is returned when a block content already has a link to a resource and another link is added.
type ErrorDuplicateResourceLink struct {
	Resource ExternalResource
}

func (e ErrorDuplicateResourceLink) Error() string {
	return "duplicate resource link: " + e.Resource.Uri
}

// AssetExistsError is returned when an asset is added to a graph that already has an asset with the same path.
type AssetExistsError struct {
	AssetPath string
}

func (e AssetExistsError) Error() string {
	return "asset already added: " + e.AssetPath
}

// AssetNotFoundError is returned when an asset is not found in a graph by path.
type AssetNotFoundError struct {
	AssetPath string
}

func (e AssetNotFoundError) Error() string {
	return "asset not found: " + e.AssetPath
}

// PageNotFoundError is returned when a page is not found in a graph by name or alias.
type PageNotFoundError struct {
	PageName string
}

func (e PageNotFoundError) Error() string {
	return "page not found: " + e.PageName
}

// PageExistsError is returned when a page is added to a graph that already has a page with the same name.
type PageExistsError struct {
	PageName string
}

func (e PageExistsError) Error() string {
	return "page already added: " + e.PageName
}
