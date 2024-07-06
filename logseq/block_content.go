package logseq

import "regexp"

type Link struct {
	Raw   string `json:"raw"`
	Url   string `json:"url"`
	Title string `json:"title"`
}

type BlockContent struct {
	Markdown      string  `json:"markdown"`
	HTML          string  `json:"html"`
	PageLinks     []*Link `json:"page_links"`
	ResourceLinks []*Link `json:"resource_links"`
}

func BlockContentFromRawSource(rawSource string) *BlockContent {
	content := BlockContent{
		Markdown:      rawSource,
		PageLinks:     findPageLinks(rawSource),
		ResourceLinks: findResourceLinks(rawSource),
	}

	return &content
}

func findPageLinks(content string) []*Link {
	pageLinkRe := regexp.MustCompile(`\[\[(.*?)\]\]`)
	pageLinks := []*Link{}

	for _, match := range pageLinkRe.FindAllStringSubmatch(content, -1) {
		raw, pageName := match[0], match[1]
		link := Link{
			Raw:   raw,
			Url:   pageName,
			Title: pageName,
		}
		pageLinks = append(pageLinks, &link)
	}

	return pageLinks
}

func findResourceLinks(content string) []*Link {
	resourceLinkRe := regexp.MustCompile(`\[(.*?)\]\((.*?)\)`)
	resourceLinks := []*Link{}

	for _, match := range resourceLinkRe.FindAllStringSubmatch(content, -1) {
		resourceTitle, resourceUrl := match[1], match[2]
		link := Link{
			Url:   resourceUrl,
			Title: resourceTitle,
		}

		resourceLinks = append(resourceLinks, &link)
	}

	return resourceLinks
}
