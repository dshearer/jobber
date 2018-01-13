DESTDIR ?= ./
WORK_DIR ?= $(abspath work)
SRC_ROOT := $(abspath ../..)
VERSION := $(shell cat ${SRC_ROOT}/version)
SRC_TARBALL = jobber-${VERSION}.tgz
SRC_TARBALL_DIR = jobber-${VERSION}
