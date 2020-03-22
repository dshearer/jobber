include ../head.mk

PKGREL := $(shell cat pkgrel)
PLATFORM = x86_64

# Alpine doesn't allow hyphens in version numbers
ALPINE_VERSION = ${shell echo ${VERSION} | sed -e 's/\-/_/' | sed -e 's/pr./rc/'}

# required by tail.mk:
PKGFILE = jobber-${ALPINE_VERSION}-r${PKGREL}.apk
PKGFILE_DEPS = \
	APKBUILD \
	jobber.initd \
	jobber.post-deinstall \
	jobber.post-install \
	jobber.post-upgrade \
	jobber.pre-deinstall \
	pkgrel
PKGFILE_VM_PATH = ${PKGFILE}
PACKAGING_SUBDIR = alpine_3.6
INSTALL_PKG_CMD = apk add --allow-untrusted ${PKGFILE}
UNINSTALL_PKG_CMD = apk remove jobber
SRC_TARBALL = jobber.tgz
SRC_TARBALL_DIR = jobber

.PHONY : pkg-local
pkg-local : ${WORK_DIR}/${SRC_TARBALL} ${PKGFILE_DEPS}
	# make package
	cp ${WORK_DIR}/${SRC_TARBALL} ./
	echo "" | abuild-keygen -a
	abuild checksum
	env "pkgrel=${PKGREL}" "pkgver=${ALPINE_VERSION}" abuild -r
	
	# copy package out of VM
	mkdir -p "${DESTDIR}/"
	find "${HOME}/packages/packaging" -name "*.apk" -exec cp '{}' "${DESTDIR}/" ';'

include ../tail.mk

test-vm :
	@echo "Cannot test on Alpine"