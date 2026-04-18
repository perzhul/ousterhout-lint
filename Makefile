.PHONY: build test lint ousterhout plugin clean

build:
	go build ./cmd/ousterhout-lint/...

test:
	go test ./...

lint:
	golangci-lint run ./...

ousterhout:
	go run ./cmd/ousterhout-lint/... ./...

plugin:
	go build -buildmode=plugin -o ousterhout.so ./golangci

clean:
	rm -f ousterhout.so
	rm -f ousterhout-lint
