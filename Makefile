GO_BUILD_LDFLAGS = -ldflags "-X 'main.Version=${TAG_VERSION}' -X 'main.GitRevision=${GIT_REVISION}'"

OS=$(shell uname)

all: sdk swig
.PHONY: all

sdk:
	cd gen/third_party && $(MAKE)

swig:
	cd gen && swig -v -go -cgo -c++ -intgosize 64 gen.i

clean:
	cd gen/third_party && $(MAKE) clean
	rm -rf bin

goformat:
	gofmt -s -w .

lint: goformat
	go install github.com/edaniels/golinters/cmd/combined
	go list -f '{{.Dir}}' ./... | grep -v gen | xargs go vet -vettool=`go env GOPATH`/bin/combined
	go list -f '{{.Dir}}' ./... | grep -v gen | xargs go run github.com/golangci/golangci-lint/cmd/golangci-lint run -v	

test:
	go test -v -coverprofile=coverage.txt -covermode=atomic ./...

.PHONY: build
build:
	mkdir -p bin && rm -rf bin; go build $(GO_BUILD_LDFLAGS) -o bin/xsens-mti-lib main.go
