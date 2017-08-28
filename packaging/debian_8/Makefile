DESTDIR ?= .
WORK_DIR ?= work
SRC_ROOT := $(abspath ../..)
SRC_TARFILE = jobber-${VERSION}.tgz
VERSION := $(shell cat ${SRC_ROOT}/version)
DOCKER_IMAGE_NAME = jobber/debian8.5_build_pkg
PKG_NAME := jobber_${VERSION}-1_amd64.deb

VM_NAME := jobber-test-debian-8
VMCMD := VBoxManage guestcontrol "${VM_NAME}" \
		--username build --password pw

.PHONY : main
main :
	@echo "Choose pkg-local or pkg-docker or pkg-vm"

.PHONY : ${WORK_DIR}/${SRC_TARFILE}
${WORK_DIR}/${SRC_TARFILE} :
	mkdir -p "${WORK_DIR}"
	make -C "${SRC_ROOT}" dist "DESTDIR=$(abspath .)/${WORK_DIR}/"

.PHONY : pkg-vm
pkg-vm : ${WORK_DIR}/${SRC_TARFILE}
	# restore "Base" snapshot and start VM
	-VBoxManage controlvm "${VM_NAME}" savestate
	VBoxManage snapshot "${VM_NAME}" restore Base
	VBoxManage startvm --type headless "${VM_NAME}"
	
	# copy Jobber source to VM
	${VMCMD} copyto --target-directory /home/build \
		$(abspath "${WORK_DIR}/${SRC_TARFILE}")
		
	# make Jobber package
	${VMCMD} run -- /bin/tar -C /home/build -xzmf \
		"/home/build/${SRC_TARFILE}"
	${VMCMD} run -- /usr/bin/make -C \
		/home/build/jobber-${VERSION}/packaging/debian_8 \
		pkg-local DESTDIR=/home/build/pkgs
	${VMCMD} run -- /bin/ls -l /home/build/pkgs
	
	# copy package out of VM
	${VMCMD} copyfrom \
		--target-directory \
		"${abspath ${DESTDIR}/${PKG_NAME}}" "/home/build/pkgs/${PKG_NAME}"
	
	# stop VM
	VBoxManage controlvm "${VM_NAME}" savestate

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

