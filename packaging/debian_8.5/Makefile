DESTDIR ?= .
WORK_DIR ?= work
SRC_ROOT := $(abspath ../..)
SRC_TARFILE = "jobber-${VERSION}.tgz"
VERSION := $(shell cat ${SRC_ROOT}/version)
DOCKER_IMAGE_NAME = jobber/debian8.5_build_pkg

.PHONY : main
main :
	@echo "Choose pkg-local or pkg-docker"

.PHONY : ${WORK_DIR}/${SRC_TARFILE}
${WORK_DIR}/${SRC_TARFILE} :
	mkdir -p "${WORK_DIR}"
	make -C "${SRC_ROOT}" dist "DESTDIR=$(abspath .)/${WORK_DIR}/"

.PHONY : pkg-docker
pkg-docker : ${WORK_DIR}/${SRC_TARFILE} docker/Dockerfile
	# bundle source
	mv "${WORK_DIR}/${SRC_TARFILE}" docker/src.tgz
	
	# run docker image
	docker build -t "${DOCKER_IMAGE_NAME}" docker
	docker run "${DOCKER_IMAGE_NAME}"
	docker cp "`docker ps -alq`:/pkgs.tar" "${DESTDIR}/"
	tar -C "${DESTDIR}/" -xmf "${DESTDIR}/pkgs.tar"
	rm -f "${DESTDIR}/pkgs.tar"

.PHONY : pkg-local
pkg-local : ${WORK_DIR}/${SRC_TARFILE}
	tar -C "${WORK_DIR}" -xf "${WORK_DIR}/${SRC_TARFILE}"
	cp "${WORK_DIR}/${SRC_TARFILE}" "${WORK_DIR}/jobber_${VERSION}.orig.tar.gz"
	cp -R debian-pkg "${WORK_DIR}/jobber-${VERSION}/debian"
	cd "${WORK_DIR}/jobber-${VERSION}" && dpkg-buildpackage -us -uc
	mkdir -p "${DESTDIR}/"
	mv "${WORK_DIR}"/*.deb "${DESTDIR}/"

.PHONY : clean
clean :
	rm -rf "${WORK_DIR}" docker/src.tgz *.deb

