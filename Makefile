DOCKER_ARCHS ?= amd64 armv7 arm64 ppc64le

include Makefile.common

DOCKER_IMAGE_NAME       ?= nagios-plugin-exporter

STRINGER        := $(FIRST_GOPATH)/bin/stringer

.PHONY: build
build: generate common-build

.PHONY: generate
generate: stringer
	$(GO) generate ./...

.PHONY: stringer
stringer: $(STRINGER)

$(STRINGER):
	$(GO) install golang.org/x/tools/cmd/stringer@latest

