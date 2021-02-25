OS=$(shell uname)

all: sdk swig
.PHONY: all

sdk:
	cd gen/third_party && $(MAKE)

swig:
	cd gen && swig -v -go -cgo -c++ -intgosize 64 mtigen.i

clean:
	cd gen/third_party && $(MAKE) clean

goformat:
	gofmt -s -w .

lint: goformat
	go list -f '{{.Dir}}' ./... | grep -v gen | xargs go run github.com/golangci/golangci-lint/cmd/golangci-lint run -v
	go get -u github.com/edaniels/golinters/cmd/combined
	go list -f '{{.Dir}}' ./... | grep -v gen | xargs go vet -vettool=`go env GOPATH`/bin/combined

test:
	go test -v -coverprofile=coverage.txt -covermode=atomic ./...
