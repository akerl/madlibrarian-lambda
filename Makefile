.PHONY: default build clean lint fmt test deps source

PACKAGE = madlibrarian-lambda
NAMESPACE = github.com/akerl
VERSION ?= $(shell git describe --tags --always --dirty --match=v* 2>/dev/null)
export GOPATH = $(CURDIR)/.gopath
BIN = $(GOPATH)/bin
BASE = $(GOPATH)/src/$(NAMESPACE)/$(PACKAGE)
GOFILES = $(shell find . -type f -name '*.go' ! -path './.*' ! -path './vendor/*')
GOPACKAGES = $(shell echo $(GOFILES) | xargs dirname | sort | uniq)

GO = go
GOFMT = gofmt
GOLINT = $(BIN)/golint
GODEP = $(BIN)/dep

build: source deps fmt lint test
	GOOS=linux GOARCH=amd64 $(GO) build -buildmode=plugin -o handler.so
	pack $(HANDLER) $(HANDLER).so $(PACKAGE).zip
	@echo "Build completed"

clean:
	rm -rf $(GOPATH) bin

lint: $(GOLINT)
	$(GOLINT) -set_exit_status $(GOPACKAGES)

fmt:
	@echo "Running gofmt on $(GOFILES)"
	@files=$$($(GOFMT) -l $(GOFILES)); if [ -n "$$files" ]; then \
		  echo "Error: '$(GOFMT)' needs to be run on:"; \
		  echo "$${files}"; \
		  exit 1; \
		  fi;

test: deps
	# TODO: Make not a noop
	#cd $(BASE) && $(GO) test $(GOPACKAGES)

deps: $(BASE) $(GODEP)
	cd $(BASE) && $(GODEP) ensure

$(BASE):
	mkdir -p $(dir $@)

source: $(BASE)
	rsync -ax --delete --exclude '.gopath' --exclude '.git' --exclude vendor $(CURDIR)/ $(BASE)

$(GOLINT): $(BASE)
	$(GO) get github.com/golang/lint/golint

$(GODEP): $(BASE)
	$(GO) get github.com/golang/dep/cmd/dep
