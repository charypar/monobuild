GOPATH:=$(shell go env GOPATH)

default: build run

# Running and testing

.PHONY: run
run: $(GOPATH)/bin/monobuild
	$(GOPATH)/bin/monobuild

test: install
	@go test ./...

# Building

build: install $(GOPATH)/bin/monobuild

~/go/bin/monobuild: ./monobuild.go
	@go install github.com/charypar/monobuild

# Dependencies

install: