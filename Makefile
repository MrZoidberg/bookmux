build:
	go build -o .bin/bookmux ./cmd/bookmux

test:
	go test ./...

lint:
	golangci-lint run

