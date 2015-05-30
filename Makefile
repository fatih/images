NO_COLOR=\033[0m
OK_COLOR=\033[0;32m
GITCOMMIT := $(shell git rev-parse --short HEAD 2>/dev/null)
GOPATH:=$(PWD)/Godeps/_workspace:$(GOPATH)
BUILD_PLATFORMS ?= -os="linux" -os="darwin" -os="windows"

all: format vet lint

deps:
	@go get -v github.com/tools/godep
	@go get -v github.com/mitchellh/gox
	@gox -build-toolchain $(BUILD_PLATFORMS)

build:
	@echo "$(OK_COLOR)==> Building the project $(NO_COLOR)"
ifndef IMAGES_VERSION
	@`which go` build
else
	@`which go` build -ldflags "-X main.Version '${IMAGES_VERSION} ($(GITCOMMIT))'"
endif

format:
	@echo "$(OK_COLOR)==> Formatting the project $(NO_COLOR)"
	@gofmt -s -w *.go
	@goimports -w *.go || true

vet:
	@echo "$(OK_COLOR)==> Running go vet $(NO_COLOR)"
	@`which go` vet .

lint:
	@echo "$(OK_COLOR)==> Running golint $(NO_COLOR)"
	@`which golint` . || true

release:
ifndef IMAGES_VERSION
	@echo "$(OK_COLOR)==> Creating development release $(NO_COLOR)"
	@gox $(BUILD_PLATFORMS) -output="out/{{.Dir}}-{{.OS}}-{{.Arch}}"
else
	@echo "$(OK_COLOR)==> Creating new release $(IMAGES_VERSION) $(NO_COLOR)"
	@gox $(BUILD_PLATFORMS) -ldflags "-X main.Version '${IMAGES_VERSION} ($(GITCOMMIT))'" -output="out/{{.Dir}}-{{.OS}}-{{.Arch}}"
endif


.PHONY: all format test vet lint
