DESTDIR ?= .
WORK_DIR ?= work
SRC_ROOT := $(abspath ../..)
SRC_TARFILE = jobber-${VERSION}.tgz
VERSION := $(shell cat ${SRC_ROOT}/version)

VM_NAME := jobber-test-ubuntu-17
VMCMD := VBoxManage guestcontrol "${VM_NAME}" \
		--username build --password pw

.PHONY : main
main :
	@echo "Choose pkg-local or pkg-vm"

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
		/home/build/jobber-${VERSION}/packaging/ubuntu_17 \
		pkg-local DESTDIR=/home/build/pkgs
	${VMCMD} run -- /bin/ls -l /home/build/pkgs
	
	# copy package out of VM
	${VMCMD} copyfrom \
		--target-directory \
		"${abspath ${DESTDIR}/jobber_${VERSION}-1_ubuntu17_amd64.deb}" \
		"/home/build/pkgs/jobber_${VERSION}-1_amd64.deb"
	
	# stop VM
	VBoxManage controlvm "${VM_NAME}" savestate

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
	rm -rf "${WORK_DIR}" *.deb

