DESTDIR ?= .
WORK_DIR ?= work
SRC_ROOT := $(abspath ../..)
SRC_TARFILE = "jobber-${VERSION}.tgz"
VERSION := $(shell cat ${SRC_ROOT}/version)
PKGREL := $(shell cat pkgrel)
PLATFORM = $(shell uname -i)
RPMFILE = jobber-${VERSION}-${PKGREL}.${PLATFORM}.rpm
DOCKER_IMAGE_NAME = jobber/centos6.7_build_rpm

.PHONY : main
main :
	@echo "Choose pkg-local or pkg-docker"

.PHONY : ${WORK_DIR}/${SRC_TARFILE}
${WORK_DIR}/${SRC_TARFILE} :
	make -C "${SRC_ROOT}" dist "DESTDIR=$(abspath .)/"
	mkdir -p "${WORK_DIR}"
	mv "./${SRC_TARFILE}" "$@"

.PHONY : pkg-docker
pkg-docker : jobber.spec ${WORK_DIR}/${SRC_TARFILE} docker/Dockerfile
	mv "${WORK_DIR}/${SRC_TARFILE}" docker/src.tgz
	docker build -t "${DOCKER_IMAGE_NAME}" docker
	docker run "${DOCKER_IMAGE_NAME}"
	docker cp "`docker ps -alq`:/rpms.tar" "${DESTDIR}/"
	tar -C "${DESTDIR}/" -xmf "${DESTDIR}/rpms.tar"
	rm -f "${DESTDIR}/rpms.tar"

.PHONY : pkg-local
pkg-local : jobber.spec ${WORK_DIR}/${SRC_TARFILE}
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
	cp "${WORK_DIR}/${SRC_TARFILE}" \
		se_policy/* \
		jobber_init "${WORK_DIR}/SOURCES/"
	
	# build RPMs
	cd "${WORK_DIR}/SPECS" && rpmbuild -bb \
		--define "_topdir ${abspath ${WORK_DIR}}" \
		--define "_pkg_version ${VERSION}" \
		--define "_pkg_release ${PKGREL}" \
		--define "_enable_debug_packages 0" \
		"$<"
	mkdir -p "${DESTDIR}/"
	cp "${WORK_DIR}/RPMS/${PLATFORM}"/*.rpm "${DESTDIR}/"

.PHONY : clean
clean :
	rm -rf "${WORK_DIR}" *.rpm docker/src.tgz

