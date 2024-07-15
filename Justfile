run:
    go run . export --env-file=.env

test:
	go test ./...

lint:
	golangci-lint run ./...
