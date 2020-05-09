include ../head.mk

# required by tail.mk:
PKGFILE = jobber_${VERSION}-1_amd64.deb
PKGFILE_DEPS = $(wildcard debian-pkg/**)
PKGFILE_VM_PATH = jobber_${VERSION}-1_amd64.deb
PACKAGING_SUBDIR = debian
INSTALL_PKG_CMD = sudo dpkg -i ${PKGFILE}
UNINSTALL_PKG_CMD = sudo dpkg -r jobber

.PHONY : pkg-local
pkg-local : ${WORK_DIR}/${SRC_TARBALL}
	cp "${WORK_DIR}/${SRC_TARBALL}" \
		"${SRC_ROOT}/../jobber_${VERSION}.orig.tar.gz"
	cp -RL debian-pkg "${SRC_ROOT}/debian"
	rm -rf "${WORK_DIR}"
	cd "${SRC_ROOT}" && dpkg-buildpackage -us -uc
	mkdir -p "${DESTDIR}/"
	mv "${SRC_ROOT}"/../*.deb "${DESTDIR}/"

include ../tail.mk
