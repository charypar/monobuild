GOPATH:=$(shell go env GOPATH)

default: build

# Running and testing

.PHONY: run
run: $(GOPATH)/bin/monobuild
	@$(GOPATH)/bin/monobuild

test: build unit-test e2e-test

e2e-test:
	@sh test/e2e.sh

e2e-test-rust: build-rust
	@sh test/e2e.sh rust

unit-test:
	@go test ./...

# Building

build: install $(GOPATH)/bin/monobuild

build-rust:
	cd rs && cargo build

$(GOPATH)/bin/monobuild: ./monobuild.go cmd/*.go diff/*.go graph/*.go manifests/*.go set/*.go cli/*.go
	@go install github.com/charypar/monobuild

# Dependencies

install: \
	$(GOPATH)/src/github.com/spf13/cobra \
	$(GOPATH)/src/github.com/bmatcuk/doublestar	

$(GOPATH)/src/github.com/spf13/cobra:
	go get github.com/spf13/cobra

$(GOPATH)/src/github.com/bmatcuk/doublestar:
	go get github.com/bmatcuk/doublestar
