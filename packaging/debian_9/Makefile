include ../head.mk

PACKAGING_SUBDIR = debian_9
PKGFILE ?= jobber_${VERSION}-1_amd64.deb
PKGFILE_DEPS = $(wildcard debian-pkg/**)
PKGFILE_VM_PATH = jobber_${VERSION}-1_amd64.deb
INSTALL_PKG_CMD = sudo dpkg -i ${PKGFILE}

.PHONY : pkg-local
pkg-local : ${WORK_DIR}/${SRC_TARFILE}
	cp "${WORK_DIR}/${SRC_TARFILE}" \
		"${SRC_ROOT}/../jobber_${VERSION}.orig.tar.gz"
	cp -R debian-pkg "${SRC_ROOT}/debian"
	cd "${SRC_ROOT}" && dpkg-buildpackage -us -uc
	mkdir -p "${DESTDIR}/"
	mv "${SRC_ROOT}"/../*.deb "${DESTDIR}/"

include ../tail.mk