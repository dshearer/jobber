SRC_ROOT = ../..
IMAGE_NAME = jobber/centos6.7_test_rpm
TEST_SCRIPTS = $(wildcard ../test_scripts/*)
VERSION := $(shell cat ${SRC_ROOT}/version)
PKGREL := $(shell cat ../../packaging/centos_6/pkgrel)
PLATFORM = x86_64
RPMFILE = jobber-${VERSION}-${PKGREL}.${PLATFORM}.rpm

VM_NAME := jobber-test-centos-6.9
VMCMD := VBoxManage guestcontrol "${VM_NAME}" \
		--username root --password password

.PHONY : test-vm
test-vm : .made_rpms_vm test_scripts.tar
	# restore "Base" snapshot and start VM
	-VBoxManage controlvm "${VM_NAME}" savestate
	VBoxManage snapshot "${VM_NAME}" restore Base
	VBoxManage startvm --type headless "${VM_NAME}"
	
	# copy RPM to VM
	${VMCMD} copyto --target-directory / \
		"$(abspath ${RPMFILE})"
		
	# install RPM
	${VMCMD} run -- /usr/bin/yum -y install "/${RPMFILE}"
	
	# copy test scripts to VM
	${VMCMD} copyto --target-directory / $(abspath test_scripts.tar)
	
	# run test scripts
	${VMCMD} run -- /bin/tar xf /test_scripts.tar
	${VMCMD} run -- /test_scripts/run_tests | tee docker_output
	@grep "platform_tests: pass" docker_output > /dev/null
	
	# stop VM
	VBoxManage controlvm "${VM_NAME}" savestate

.PHONY : test-docker
test-docker : .made_rpms test_scripts.tar
	docker build -t "${IMAGE_NAME}" .
	docker run "${IMAGE_NAME}" | tee docker_output
	@grep "platform_tests: pass" docker_output > /dev/null

test_scripts.tar : ${TEST_SCRIPTS}
	tar -C .. -cf $@ test_scripts

.made_rpms :
	make -C "${SRC_ROOT}/packaging/centos_6" pkg-docker "DESTDIR=$(abspath .)/"
	touch "$@"

.made_rpms_vm :
	make -C "${SRC_ROOT}/packaging/centos_6" pkg-vm "DESTDIR=$(abspath .)/"
	touch "$@"

.PHONY : clean
clean :
	rm -f *.rpm test_scripts.tar docker_output .made_rpms .made_rpms_vm
	make -C "${SRC_ROOT}/packaging/centos_6" clean
