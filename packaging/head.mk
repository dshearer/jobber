DESTDIR ?= ./
WORK_DIR ?= $(abspath work)
SRC_ROOT := $(abspath ../..)
VERSION := $(shell cat ${SRC_ROOT}/version)
SRC_TARFILE = jobber-${VERSION}.tgz
ROBOT_TESTS_DIR = ${SRC_ROOT}/platform_tests/robot_tests
ROBOT_TESTS = $(wildcard ${ROBOT_TESTS_DIR}/*)

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
