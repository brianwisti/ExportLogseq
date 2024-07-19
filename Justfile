public:
    go run . export --env-file=.env

all:
	go run . export --env-file=.env --selected-pages=all
test:
	go test -cover ./...

lint:
	golangci-lint run ./...
