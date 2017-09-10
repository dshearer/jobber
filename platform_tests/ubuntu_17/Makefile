SRC_ROOT = ../..
IMAGE_NAME = jobber/ubuntu16.04_test_pkg
ROBOT_TESTS = $(wildcard ${SRC_ROOT}/platform_tests/robot_tests/*)
VERSION := $(shell cat ${SRC_ROOT}/version)
PKG_NAME := jobber_${VERSION}-1_ubuntu17_amd64.deb

VM_NAME := jobber-test-ubuntu-17
VMCMD := VBoxManage guestcontrol "${VM_NAME}" \
		--username root --password pw

.PHONY : test-vm
test-vm : .made_pkgs_vm robot_tests.tar
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
	${VMCMD} copyto --target-directory / $(abspath robot_tests.tar)
	
	# run test scripts
	${VMCMD} run -- /bin/tar xf /robot_tests.tar
	-${VMCMD} run -- /usr/local/bin/robot /robot_tests/test.robot > testlog.txt
	
	# retrieve test reports
	mkdir -p "${DESTDIR}test_report"
	${VMCMD} copyfrom --target-directory \
		"$(abspath ${DESTDIR}test_report)/log.html" /log.html
	${VMCMD} copyfrom --target-directory \
		"$(abspath ${DESTDIR}test_report)/report.html" /report.html
	
	# stop VM
	VBoxManage controlvm "${VM_NAME}" savestate
	
	# finish up
	@cat testlog.txt
	@egrep '^Test[[:space:]]+[|] PASS [|]$$' testlog.txt

robot_tests.tar : ${ROBOT_TESTS}
	tar -C .. -cf $@ robot_tests

.made_pkgs_vm :
	make -C "${SRC_ROOT}/packaging/ubuntu_17" pkg-vm "DESTDIR=$(abspath .)/"
	touch "$@"

.PHONY : clean
clean :
	rm -rf *.deb robot_tests.tar testlog.txt .made_pkgs_vm \
		"${DESTDIR}test_report"
	make -C "${SRC_ROOT}/packaging/ubuntu_17" clean
