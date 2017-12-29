DESTDIR ?= ./
SRC_ROOT := $(abspath ../..)
DOCKER_IMAGE_NAME = jobber/alpine3.6_build_pkg
PLATFORM = $(shell uname -i)
PKGREL = $(shell cat pkgrel)
VERSION = $(shell cat "${SRC_ROOT}/version")

.PHONY : main
main :
	@echo "Choose pkg-local or pkg-docker"

.PHONY : pkg-docker
pkg-docker : jobber.tgz
	# run VM
	cp jobber.tgz docker/
	docker build -t "${DOCKER_IMAGE_NAME}" docker
	docker run "${DOCKER_IMAGE_NAME}"
	
	# extract package from VM
	docker cp "`docker ps -alq`:/output/pkgs.tgz" "${DESTDIR}"
	tar -C "${DESTDIR}" -xmzf "${DESTDIR}pkgs.tgz"
	rm -f "${DESTDIR}pkgs.tgz"

.PHONY : pkg-local
pkg-local : pkgrel jobber.tgz
	# make package
	ln -s $(abspath jobber.tgz) /var/cache/distfiles/jobber.tgz
	abuild-keygen -a
	abuild checksum
	env "pkgrel=${PKGREL}" "pkgver=${VERSION}" abuild -r
	
	# make pkgs.tgz
	find "${HOME}/packages/packaging" -name "*.apk" -exec cp '{}' "${DESTDIR}" ';'
	
jobber.tgz :
	# make source package
	make -C "${SRC_ROOT}" "DESTDIR=$(abspath .)/" dist DIST_PKG_NAME=jobber

.PHONY : clean
clean :
	rm -f jobber.tgz docker/jobber.tgz "${DESTDIR}"*.apk

.PHONY : clean-docker
clean-docker :
	-abuild clean
	-abuild cleanoldpkg
	-abuild cleanpkg
	-abuild cleancache
