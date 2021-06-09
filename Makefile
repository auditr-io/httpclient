.PHONY: build test clean

build:
	export GO111MODULE=on
	go build -ldflags="-s -w" -o bin/httpclient *.go

test:
	ENV_PATH=$(shell pwd)/testdata/dotenv go test -timeout 30s -v ./...

clean:
	rm -rf ./bin ./vendor Gopkg.lock
	rm -rf dist
