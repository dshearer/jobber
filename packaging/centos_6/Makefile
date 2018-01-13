include ../head.mk

PKGREL := $(shell cat pkgrel)
PLATFORM = x86_64

# RPMs don't allow hyphens in version numbers
RPM_VERSION = ${shell echo ${VERSION} | sed -e 's/\-/_/'}

# required by tail.mk:
PKGFILE = jobber-${RPM_VERSION}-${PKGREL}.el6.centos.${PLATFORM}.rpm
PKGFILE_DEPS = jobber.spec
PKGFILE_VM_PATH = jobber-${RPM_VERSION}-${PKGREL}.${PLATFORM}.rpm
PACKAGING_SUBDIR = centos_6
INSTALL_PKG_CMD = sudo yum install -y ${PKGFILE}
UNINSTALL_PKG_CMD = sudo yum remove -y jobber
SRC_TARBALL = jobber-${RPM_VERSION}.tgz
SRC_TARBALL_DIR = jobber-${RPM_VERSION}

.PHONY : pkg-local
pkg-local : jobber.spec ${WORK_DIR}/${SRC_TARBALL}
	# make RPM tree
	mkdir -p "${WORK_DIR}" \
		 "${WORK_DIR}/BUILD" \
                 "${WORK_DIR}/RPMS" \
                 "${WORK_DIR}/RPMS/${PLATFORM}" \
                 "${WORK_DIR}/SOURCES" \
                 "${WORK_DIR}/SPECS" \
                 "${WORK_DIR}/SRPMS"
	cp "$<" "${WORK_DIR}/SPECS"
	
	# copy sources
	cp "${WORK_DIR}/${SRC_TARBALL}" \
		se_policy/* \
		jobber_init "${WORK_DIR}/SOURCES/"
	
	# build RPMs
	cd "${WORK_DIR}/SPECS" && rpmbuild -bb \
		--define "_topdir ${abspath ${WORK_DIR}}" \
		--define "_pkg_version ${RPM_VERSION}" \
		--define "_pkg_release ${PKGREL}" \
		--define "_enable_debug_packages 0" \
		"$<"
	mkdir -p "${DESTDIR}"
	find "${WORK_DIR}/RPMS" -name '*.rpm' | xargs cp -t "${DESTDIR}"

include ../tail.mk