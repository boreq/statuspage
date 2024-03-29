VERSION = `git rev-parse HEAD`
DATE = `date --iso-8601=seconds`
LDFLAGS =  -X github.com/boreq/statuspage-backend/main/commands.buildCommit=$(VERSION)
LDFLAGS += -X github.com/boreq/statuspage-backend/main/commands.buildDate=$(DATE)

all: build

build:
	mkdir -p build
	go build -ldflags "$(LDFLAGS)" -o ./build/statuspage-backend ./main

doc:
	@echo "http://localhost:6060/pkg/github.com/boreq/statuspage-backend/"
	godoc -http=:6060

test:
	go test ./...

test-verbose:
	go test -v ./...

test-short:
	go test -short ./...

bench:
	go test -v -run=XXX -bench=. ./...

clean:
	rm -rf ./build

tools:
	go install github.com/tinylib/msgp

generate:
	go generate ./...

.PHONY: all build doc test test-verbose test-short bench clean tools generate
