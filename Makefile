PROJECT=cores-go
PACKAGE=github.com/catalyst-zero/$(PROJECT)

PWD := $(shell pwd)

BUILD_PATH := $(PWD)/.gobuild

C0_PATH := $(BUILD_PATH)/src/github.com/catalyst-zero/

BIN=producer consumer

.PHONY=clean run-test get-deps update-deps

GOPATH := $(BUILD_PATH)

SOURCE=$(wildcard *.go)

all: get-deps $(BIN)

clean:
	rm -rf $(BUILD_PATH) $(BIN)

get-deps: .gobuild

.gobuild:
	mkdir -p $(C0_PATH)
	cd "$(C0_PATH)" && ln -s ../../../.. $(PROJECT)

	#
	# Fetch public dependencies via `go get`
	GOPATH=$(GOPATH) go get -d -v $(PACKAGE)

	#
	# Fetch deployment dependencies
	#GOPATH=$(GOPATH) go get github.com/kr/godep

	#
	# Build test packages (we only want those two, so we use `-d` in go get)
	#GOPATH=$(GOPATH) go get -v github.com/onsi/gomega
	#GOPATH=$(GOPATH) go get -v github.com/onsi/ginkgo

$(BIN): $(SOURCE)
	GOBIN=$(PWD) GOPATH=$(GOPATH) go install $(PACKAGE)/example/$(BIN)

run-tests:
	GOPATH=$(GOPATH) go test ./...

fmt:
	gofmt -l -w .
