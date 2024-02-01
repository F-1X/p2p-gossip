build:
	@go build -o bin/gossip

run: build
	@./bin/gossip

test:
	go test -v ./...