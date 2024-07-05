package logseq

import "regexp"

type Link struct {
	Url   string `json:"url"`
	Title string `json:"title"`
}

type BlockContent struct {
	Markdown      string `json:"markdown"`
	PageLinks     []Link `json:"page_links"`
	ResourceLinks []Link `json:"resource_links"`
}

func BlockContentFromRawSource(rawSource string) BlockContent {
	return BlockContent{
		Markdown:      rawSource,
		PageLinks:     findPageLinks(rawSource),
		ResourceLinks: findResourceLinks(rawSource),
	}
}

func findPageLinks(content string) []Link {
	pageLinkRe := regexp.MustCompile(`\[\[(.*?)\]\]`)
	pageLinks := []Link{}

	for _, match := range pageLinkRe.FindAllStringSubmatch(content, -1) {
		pageName := match[1]
		pageLinks = append(pageLinks, Link{
			Url:   pageName,
			Title: pageName,
		})
	}

	return pageLinks
}

func findResourceLinks(content string) []Link {
	resourceLinkRe := regexp.MustCompile(`\[(.*?)\]\((.*?)\)`)
	resourceLinks := []Link{}

	for _, match := range resourceLinkRe.FindAllStringSubmatch(content, -1) {
		resourceTitle, resourceUrl := match[1], match[2]
		resourceLinks = append(resourceLinks, Link{
			Url:   resourceUrl,
			Title: resourceTitle,
		})
	}

	return resourceLinks
}
