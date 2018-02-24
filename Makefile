.PHONY: default build clean lint fmt test deps init update

PACKAGE = madlibrarian-lambda
NAMESPACE = github.com/akerl
VERSION ?= $(shell git describe --tags --always --dirty --match=v* 2>/dev/null)
export GOPATH = $(CURDIR)/.gopath
BIN = $(GOPATH)/bin
BASEDIR = $(GOPATH)/src/$(NAMESPACE)
BASE = $(BASEDIR)/$(PACKAGE)
GOFILES = $(shell find . -type f -name '*.go' ! -path './.*' ! -path './vendor/*')

GO = go
GOFMT = gofmt
GOX = $(BIN)/gox
GOLINT = $(BIN)/golint
GODEP = $(BIN)/dep

build: $(BASE) deps $(GOX) fmt lint test
	cd $(BASE) && $(GOX) \
		-ldflags '-X $(NAMESPACE)/$(PACKAGE)/cmd.Version=$(VERSION)' \
		-gocmd="$(GO)" \
		-output="bin/$(PACKAGE)_{{.OS}}" \
		-os="linux" \
		-arch="amd64"
	@echo "Build completed"

clean:
	rm -rf $(GOPATH) bin

lint: $(GOLINT)
	$(GOLINT) -set_exit_status ./...

fmt:
	@echo "Running gofmt on $(GOFILES)"
	@files=$$($(GOFMT) -l $(GOFILES)); if [ -n "$$files" ]; then \
		  echo "Error: '$(GOFMT)' needs to be run on:"; \
		  echo "$${files}"; \
		  exit 1; \
		  fi;

test: deps
	cd $(BASE) && $(GO) test ./...

init: $(BASE) $(GODEP)
	cd $(BASE) && $(GODEP) init
	cp $(BASE)/Gopkg.{lock,toml} ./

update: $(BASE) $(GODEP)
	cd $(BASE) && $(GODEP) ensure -update
	cp $(BASE)/Gopkg.{lock,toml} ./

deps: $(BASE) $(GODEP)
	cd $(BASE) && $(GODEP) ensure

$(BASEDIR):
	mkdir -p $(BASEDIR)

$(BASE): $(BASEDIR)
	ln -s $(BASE) $(CURDIR)

$(GOLINT): $(BASE)
	$(GO) get github.com/golang/lint/golint

$(GOX): $(BASE)
	$(GO) get github.com/mitchellh/gox

$(GODEP): $(BASE)
	$(GO) get github.com/golang/dep/cmd/dep
