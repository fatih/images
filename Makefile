NO_COLOR=\033[0m
OK_COLOR=\033[0;32m
ERR_COLOR=\033[0;31m
GITCOMMIT := $(shell git rev-parse --short HEAD 2>/dev/null)
GOPATH:=$(PWD)/Godeps/_workspace:$(GOPATH)

all: build

build:
	@echo "$(OK_COLOR)==> Building the project $(NO_COLOR)"
ifndef IMAGES_VERSION
	@`which go` build -v -o bin/images
else
	@`which go` build -v -ldflags "-X main.Version '${IMAGES_VERSION} ($(GITCOMMIT))'" -o bin/images
endif

release: check_goxc clean
ifdef IMAGES_VERSION
	@echo "$(OK_COLOR)==> Creating new release $(IMAGES_VERSION) $(NO_COLOR)"
	@goxc -arch "386 amd64" -os="linux windows darwin" -d "out" -pv $(IMAGES_VERSION) -build-ldflags="-X main.Version '${IMAGES_VERSION} ($(GITCOMMIT))'" -n images -q
	@rm -rf debian/
else
	@echo "$(ERR_COLOR)Please set IMAGES_VERRSION environment variable to create a release $(NO_COLOR)"
endif

check_goxc:
	@echo "$(OK_COLOR)==> Checking goxc availability $(NO_COLOR)"
	@which goxc > /dev/null

clean:
	@echo "$(OK_COLOR)==> Cleaning output directories $(NO_COLOR)"
	@rm -rf out/
	@rm -rf debian/
	@rm -rf images

.PHONY: all clean
