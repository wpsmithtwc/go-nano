.PHONY: examples doc fmt lint test vet

# Prepend our _vendor directory to the system GOPATH
# so that import path resolution will prioritize
# our third party snapshots.
GOPATH := ${PWD}/_vendor:${GOPATH}
export GOPATH

default: test

examples:
		go build -v -o upper ./examples/upper/main.go
		go build -v -o demo ./examples/demo/main.go

doc:
		godoc -http=:6060 -index

# http://golang.org/cmd/go/#hdr-Run_gofmt_on_package_sources
fmt:
		go fmt ./...

# https://github.com/golang/lint
# go get github.com/golang/lint/golint
lint:
		golint .

test:
		go test ./...

# http://godoc.org/code.google.com/p/go.tools/cmd/vet
# go get code.google.com/p/go.tools/cmd/vet
vet:
		go vet ./...