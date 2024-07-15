run:
    go run . export ~/my-logseq-brain ~/www/rgb-logseq-hugo

test:
	go test ./...

lint:
	golangci-lint run ./...
