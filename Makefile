DOCKER_ARCHS ?= amd64 armv7 arm64 ppc64le

include Makefile.common

DOCKER_IMAGE_NAME       ?= nagios-plugin-exporter
