GOPATH:=$(shell go env GOPATH)

default: build

# Running and testing

.PHONY: run
run: $(GOPATH)/bin/monobuild
	@$(GOPATH)/bin/monobuild

test: install
	@go test ./...

# Building

build: install $(GOPATH)/bin/monobuild

~/go/bin/monobuild: ./monobuild.go cmd/*.go diff/*.go graph/*.go set/*.go
	@go install github.com/charypar/monobuild

# Dependencies

install: \
	$(GOPATH)/src/github.com/spf13/cobra \
	$(GOPATH)/src/github.com/bmatcuk/doublestar	

$(GOPATH)/src/github.com/spf13/cobra:
	go get github.com/spf13/cobra

$(GOPATH)/src/github.com/bmatcuk/doublestar:
	go get github.com/bmatcuk/doublestar