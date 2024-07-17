# Logseq Custom Export

A Go executable that exports my [Logseq][logseq] graph to a custom JSON format or a Hugo site.

[logseq]: https://logseq.com

## Caveats

This is another Go learning project. Look at it for ideas, but maybe don't actually use it yet. Later. I promise.

## Scratchpad

- pages get exported `journals/` or `pages/` depending on where they're found.
- Set `hoist-namespace` property to true for namespaces you want at the top level; say for example `post/`; that page and its subpages will be hoisted up to the main content level
