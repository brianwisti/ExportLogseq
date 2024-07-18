run:
    go run . export --env-file=.env

test:
	go test -cover ./...

lint:
	golangci-lint run ./...
