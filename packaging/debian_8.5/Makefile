DESTDIR ?= .
WORK_DIR ?= work
SRC_ROOT := $(abspath ../..)
SRC_TARFILE = "jobber_${VERSION}.orig.tar.gz"
VERSION := $(shell cat ${SRC_ROOT}/version)
DOCKER_IMAGE_NAME = jobber/centos6.7_build_rpm

.PHONY : main
main :
	@echo "Choose pkg-local or pkg-docker"

.PHONY : ${WORK_DIR}/${SRC_TARFILE}
${WORK_DIR}/${SRC_TARFILE} :
	make -C "${SRC_ROOT}" dist "DESTDIR=$(abspath .)/"
	mkdir -p "${WORK_DIR}"
	mv "jobber-${VERSION}.tgz" "$@"

.PHONY : pkg-docker
pkg-docker : ${WORK_DIR}/${SRC_TARFILE} docker/Dockerfile
	mv "${WORK_DIR}/${SRC_TARFILE}" docker/src.tgz
	docker build -t "${DOCKER_IMAGE_NAME}" docker
	docker run "${DOCKER_IMAGE_NAME}"
	docker cp "`docker ps -alq`:/pkgs.tar" "${DESTDIR}/"
	tar -C "${DESTDIR}/" -xmf "${DESTDIR}/pkgs.tar"
	rm -f "${DESTDIR}/pkgs.tar"

.PHONY : pkg-local
pkg-local : ${WORK_DIR}/${SRC_TARFILE}
	tar -C "${WORK_DIR}" -xf "${WORK_DIR}/${SRC_TARFILE}"
	cp -R debian-pkg "${WORK_DIR}/jobber-${VERSION}/debian"
	cd "${WORK_DIR}/jobber-${VERSION}" && dpkg-buildpackage -us -uc
	mv "${WORK_DIR}"/*.deb "${DESTDIR}/"

.PHONY : clean
clean :
	rm -rf "${WORK_DIR}" docker/src.tgz *.deb

