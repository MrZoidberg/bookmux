build:
	go build -o .bin/bookmux ./cmd/bookmux

test:
	go test ./...

lint:
	golangci-lint run

tests-e2e:
	go test -v -tags=e2e ./e2e
