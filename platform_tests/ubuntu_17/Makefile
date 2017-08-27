SRC_ROOT = ../..
IMAGE_NAME = jobber/ubuntu16.04_test_pkg
TEST_SCRIPTS := $(wildcard ../test_scripts/*)
VERSION := $(shell cat ${SRC_ROOT}/version)
PKG_NAME := jobber_${VERSION}-1_amd64.deb

VM_NAME := jobber-test-ubuntu-17
VMCMD := VBoxManage guestcontrol "${VM_NAME}" \
		--username root --password pw

.PHONY : test-vm
test-vm : .made_pkgs_vm test_scripts.tar
	# restore "Base" snapshot and start VM
	-VBoxManage controlvm "${VM_NAME}" savestate
	VBoxManage snapshot "${VM_NAME}" restore Base
	VBoxManage startvm --type headless "${VM_NAME}"
	
	# copy package to VM
	${VMCMD} copyto --target-directory / \
		"$(abspath ${PKG_NAME})"
		
	# install package
	${VMCMD} run -- /usr/bin/dpkg -i "/${PKG_NAME}"
	
	# make sure Jobber is running
	${VMCMD} run -- /bin/systemctl status jobber
	
	# copy test scripts to VM
	${VMCMD} copyto --target-directory / $(abspath test_scripts.tar)
	
	# run test scripts
	${VMCMD} run -- /bin/tar xf /test_scripts.tar
	${VMCMD} run -- /test_scripts/run_tests | tee docker_output
	@grep "platform_tests: pass" docker_output > /dev/null
	
	# stop VM
	VBoxManage controlvm "${VM_NAME}" savestate

.PHONY : test
test : .made_pkgs test_scripts.tar
	docker build -t "${IMAGE_NAME}" .
	docker run "${IMAGE_NAME}" | tee docker_output
	@grep "platform_tests: pass" docker_output > /dev/null

test_scripts.tar : ${TEST_SCRIPTS}
	tar -C .. -cf $@ test_scripts

.made_pkgs_vm :
	make -C "${SRC_ROOT}/packaging/debian_8" pkg-vm "DESTDIR=$(abspath .)/"
	touch "$@"

.PHONY : clean
clean :
	rm -f *.deb test_scripts.tar docker_output .made_pkgs_vm
	make -C "${SRC_ROOT}/packaging/debian_8" clean
