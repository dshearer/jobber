include ../head.mk

PACKAGING_SUBDIR = ubuntu_16
PKGFILE ?= jobber_${VERSION}-1_amd64.deb
PKGFILE_DEPS = $(wildcard debian-pkg/**)
PKGFILE_VM_PATH = jobber_${VERSION}-1_amd64.deb
INSTALL_PKG_CMD = sudo dpkg -i ${PKGFILE}

.PHONY : pkg-local
pkg-local : ${WORK_DIR}/${SRC_TARFILE}
	tar -C "${WORK_DIR}" -xf "${WORK_DIR}/${SRC_TARFILE}"
	cp "${WORK_DIR}/${SRC_TARFILE}" \
		"${WORK_DIR}/jobber_${VERSION}.orig.tar.gz"
	cp -R debian-pkg "${WORK_DIR}/jobber-${VERSION}/debian"
	cd "${WORK_DIR}/jobber-${VERSION}" && dpkg-buildpackage -us -uc
	mkdir -p "${DESTDIR}/"
	mv "${WORK_DIR}"/*.deb "${DESTDIR}/"

include ../tail.mk