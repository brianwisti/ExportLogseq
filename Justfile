
test:
	go test export-logseq/graph

lint:
	golangci-lint run ./. ./graph ./logseq ./hugo
