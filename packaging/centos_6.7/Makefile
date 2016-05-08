WORK_DIR ?= work
SRC_ROOT = ../..
VERSION = $(shell cat ${SRC_ROOT}/version)
PKGREL = $(shell cat pkgrel)
TARFILE = "jobber-${VERSION}.tgz"
PLATFORM = $(shell uname -i)
RPMFILE = "jobber-${VERSION}-${PKGREL}.${PLATFORM}.rpm"

.PHONY : build
build : ${WORK_DIR}/.done

${WORK_DIR}/${TARFILE} :
	make -C "${SRC_ROOT}" dist
	mkdir -p "${WORK_DIR}"
	mv "${SRC_ROOT}/${TARFILE}" "$@"

${WORK_DIR}/.done : jobber.spec ${WORK_DIR}/${TARFILE}
	# make RPM tree
	mkdir -p "${dir $@}" \
		 "${dir $@}BUILD" \
                 "${dir $@}RPMS" \
                 "${dir $@}RPMS/${PLATFORM}" \
                 "${dir $@}SOURCES" \
                 "${dir $@}SPECS" \
                 "${dir $@}SRPMS"
	cp "$<" "${dir $@}SPECS"
	
	# copy sources
	cp "${WORK_DIR}/${TARFILE}" jobber_init "${dir $@}SOURCES/"
	
	# build RPM
	cd "${dir $@}SPECS" && rpmbuild -bb \
		--define "_topdir ${abspath ${dir $@}}" \
		--define "_pkg_version ${VERSION}" \
		--define "_pkg_release ${PKGREL}" \
		--define "_enable_debug_packages 0" \
		"$<"
	cp "${dir $@}RPMS/${PLATFORM}/${RPMFILE}" ./
	touch "$@"

.PHONY : clean
clean :
	rm -rf "${WORK_DIR}" "${RPMFILE}"

