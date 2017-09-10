SRC_ROOT = ../..
IMAGE_NAME = jobber/debian8.5_test_pkg
TEST_SCRIPTS = $(wildcard ${SRC_ROOT}/platform_tests/test_scripts/*)
ROBOT_TESTS = $(wildcard ${SRC_ROOT}/platform_tests/robot_tests/*)
VERSION := $(shell cat ${SRC_ROOT}/version)
PKG_NAME := jobber_${VERSION}-1_debian8_amd64.deb

VM_NAME := jobber-test-debian-8
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

.PHONY : test-docker
test-docker : .made_pkgs test_scripts.tar
	docker build -t "${IMAGE_NAME}" .
	docker run "${IMAGE_NAME}" | tee docker_output
	@grep "platform_tests: pass" docker_output > /dev/null

robot_tests.tar : ${ROBOT_TESTS}
	tar -C .. -cf $@ robot_tests

test_scripts.tar : ${TEST_SCRIPTS}
	tar -C .. -cf $@ test_scripts

.made_pkgs :
	make -C "${SRC_ROOT}/packaging/debian_8" pkg-docker "DESTDIR=$(abspath .)/"
	touch "$@"

.made_pkgs_vm :
	make -C "${SRC_ROOT}/packaging/debian_8" pkg-vm "DESTDIR=$(abspath .)/"
	touch "$@"

.PHONY : clean
clean :
	rm -rf *.deb test_scripts.tar robot_tests.tar testlog.txt \
		.made_pkgs .made_pkgs_vm "${DESTDIR}test_report"
	make -C "${SRC_ROOT}/packaging/debian_8" clean
