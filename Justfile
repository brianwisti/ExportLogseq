test:
	go test export-logseq/logseq

lint:
	golangci-lint run ./. ./logseq