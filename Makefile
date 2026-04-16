build:
	go build -o .bin/bookmux ./cmd/bookmux

test:
	go test ./...

coverage:
	go test -covermode=atomic -coverprofile=coverage.out ./...

lint:
	golangci-lint run

tests-e2e:
	go test -v -tags=e2e ./e2e
