DESTDIR ?= ./
WORK_DIR ?= $(abspath work)
SRC_ROOT := $(abspath ../..)
VERSION := $(shell cat ${SRC_ROOT}/version)
SRC_TARFILE = jobber-${VERSION}.tgz

# Must be defined by includers:
#	Variables:
# 		PKGFILE
# 		PKGFILE_DEPS
# 		PKGFILE_VM_PATH
#   		PACKAGING_SUBDIR
#		INSTALL_PKG_CMD
#		DOCKER_IMAGE_NAME
#
#	Rules:
#		pkg-local
