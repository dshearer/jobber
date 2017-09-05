DESTDIR ?= .
WORK_DIR ?= work
SRC_ROOT := $(abspath ../..)
SRC_TARFILE = "jobber-${VERSION}.tgz"
VERSION := $(shell cat ${SRC_ROOT}/version)
PKGREL := $(shell cat pkgrel)
PLATFORM = x86_64
DOCKER_IMAGE_NAME = jobber/centos6.7_build_rpm

VM_NAME := jobber-test-centos-6.9
VMCMD := VBoxManage guestcontrol "${VM_NAME}" \
		--username build --password pw

.PHONY : main
main :
	@echo "Choose pkg-local or pkg-docker or pkg-vm"

.PHONY : ${WORK_DIR}/${SRC_TARFILE}
${WORK_DIR}/${SRC_TARFILE} :
	make -C "${SRC_ROOT}" dist "DESTDIR=$(abspath .)/"
	mkdir -p "${WORK_DIR}"
	mv "./${SRC_TARFILE}" "$@"

.PHONY : pkg-vm
pkg-vm : jobber.spec ${WORK_DIR}/${SRC_TARFILE}
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
		/home/build/jobber-${VERSION}/packaging/centos_6 \
		pkg-local DESTDIR=/home/build/rpms
	${VMCMD} run -- /bin/ls -l /home/build/rpms
	
	# copy package out of VM
	${VMCMD} copyfrom \
		--target-directory \
		${abspath ${DESTDIR}/jobber-${VERSION}-${PKGREL}.el6.centos.${PLATFORM}.rpm} \
		/home/build/rpms/jobber-${VERSION}-${PKGREL}.${PLATFORM}.rpm
	
	# stop VM
	VBoxManage controlvm "${VM_NAME}" savestate

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
	find "${WORK_DIR}/RPMS" -name '*.rpm' | xargs cp -t "${DESTDIR}"

.PHONY : clean
clean :
	rm -rf "${WORK_DIR}" *.rpm docker/src.tgz

