SRC_ROOT = ../..
IMAGE_NAME = jobber/centos6.7_test_rpm
TEST_SCRIPTS = $(wildcard ${SRC_ROOT}/platform_tests/test_scripts/*)
ROBOT_TESTS = $(wildcard ${SRC_ROOT}/platform_tests/robot_tests/*)
VERSION := $(shell cat ${SRC_ROOT}/version)
PKGREL := $(shell cat ../../packaging/centos_6/pkgrel)
PLATFORM = x86_64
RPMFILE = jobber-${VERSION}-${PKGREL}.el6.centos.${PLATFORM}.rpm

VM_NAME := jobber-test-centos-6.9
VMCMD := VBoxManage guestcontrol "${VM_NAME}" \
		--username root --password password

.PHONY : test-vm
test-vm : .made_rpms_vm robot_tests.tar
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
	${VMCMD} copyto --target-directory / $(abspath robot_tests.tar)
	
	# run test scripts
	${VMCMD} run -- /bin/tar xf /robot_tests.tar
	-${VMCMD} run -- /usr/bin/robot /robot_tests/test.robot > testlog.txt
	
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
test-docker : .made_rpms test_scripts.tar
	docker build -t "${IMAGE_NAME}" .
	docker run "${IMAGE_NAME}" | tee docker_output
	@grep "platform_tests: pass" docker_output > /dev/null

robot_tests.tar : ${ROBOT_TESTS}
	tar -C .. -cf $@ robot_tests

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
	rm -rf *.rpm test_scripts.tar robot_tests.tar testlog.txt \
		.made_rpms .made_rpms_vm "${DESTDIR}test_report"
	make -C "${SRC_ROOT}/packaging/centos_6" clean
