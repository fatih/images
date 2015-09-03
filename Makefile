NO_COLOR=\033[0m
OK_COLOR=\033[0;32m
ERR_COLOR=\033[0;31m
GITCOMMIT := $(shell git rev-parse --short HEAD 2>/dev/null)
GOPATH:=$(PWD)/Godeps/_workspace:$(GOPATH)

all: build

install:
	@echo "$(OK_COLOR)==> Installing to /usr/local/bin $(NO_COLOR)"
	@cp bin/images /usr/local/bin

build: check_gb
	@echo "$(OK_COLOR)==> Building the project $(NO_COLOR)"
ifndef IMAGES_VERSION
	@`which gb` build
	@echo "$(OK_COLOR)==> Binary installed to bin/ $(NO_COLOR)"
else
	@`which go` build -v -ldflags "-X main.Version '${IMAGES_VERSION} ($(GITCOMMIT))'" -o bin/images
endif

release: check_gb clean
ifdef IMAGES_VERSION
	@echo "$(OK_COLOR)==> Creating new release $(IMAGES_VERSION) $(NO_COLOR)"
	@env GOOS=linux GOARCH=amd64 gb build -ldflags="-X main.Version=${IMAGES_VERSION}"
	@env GOOS=windows GOARCH=amd64 gb build -ldflags="-X main.Version=${IMAGES_VERSION}"
	@env GOOS=darwin GOARCH=amd64 gb build -ldflags="-X main.Version=${IMAGES_VERSION}"

	@echo "$(OK_COLOR)==> Constructing archives"
	@mkdir -p out/
	@zip images_${IMAGES_VERSION}_darwin_amd64.zip bin/images
	@zip images_${IMAGES_VERSION}_linux_amd64.zip bin/images-linux-amd64
	@zip images_${IMAGES_VERSION}_windows_amd64.zip bin/images-windows-amd64.exe
	@mv *.zip out/
	
else
	@echo "$(ERR_COLOR)Please set IMAGES_VERRSION environment variable to create a release $(NO_COLOR)"
endif

check_gb:
	@echo "$(OK_COLOR)==> Checking gb binary $(NO_COLOR)"
	@which gb > /dev/null

clean:
	@echo "$(OK_COLOR)==> Cleaning output directories $(NO_COLOR)"
	@rm -rf out/
	@rm -rf bin/

.PHONY: all clean
