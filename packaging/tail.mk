VAGRANT_SSH = vagrant ssh --no-tty -c

.PHONY : main
main :
	@echo "Choose pkg-local or pkg-vm or test-vm"

.PHONY : pkg-vm
pkg-vm : .vm-is-pristine ${DESTDIR}${PKGFILE}
	# stop VM
	vagrant suspend

.vm-is-pristine : .vm-is-running
	# restore "Base" snapshot and start VM
	-vagrant suspend
	vagrant snapshot restore Base
	touch $@
	
.vm-is-running :
	(vagrant snapshot list | grep Base >/dev/null) || \
		(vagrant up && vagrant reload && vagrant suspend && \
			vagrant snapshot save Base)
	touch $@

${DESTDIR}${PKGFILE} : Vagrantfile ${WORK_DIR}/${SRC_TARFILE} \
		${PKGFILE_DEPS} .vm-is-running
	
	rm -f .vm-is-pristine
	
	# make sure VM is running
	(vagrant status | grep running >/dev/null) || vagrant up
	
	# copy Jobber source to VM
	vagrant scp "${WORK_DIR}/${SRC_TARFILE}" ":${SRC_TARFILE}"
	${VAGRANT_SSH} "tar -xzmf ${SRC_TARFILE}"
	
	# make Jobber package
	${VAGRANT_SSH} "mkdir -p work && \
		mv ${SRC_TARFILE} work/${SRC_TARFILE} && \
		mkdir -p dest && \
		make -C jobber-${VERSION}/packaging/${PACKAGING_SUBDIR} \
		pkg-local DESTDIR=~/dest/ WORK_DIR=~/work"
	
	# copy package out of VM
	vagrant scp :dest/${PKGFILE_VM_PATH} "${DESTDIR}${PKGFILE}"
	
	touch "$@"

.PHONY : test-vm
test-vm : .vm-is-pristine test-vm-dev
	# stop VM
	vagrant suspend

.PHONY : test-vm-dev
test-vm-dev : .vm-is-running ${DESTDIR}${PKGFILE} platform_tests.tar
	rm -f .vm-is-pristine
	
	# make sure VM is running
	(vagrant status | grep running >/dev/null) || vagrant up
	
	# install package
	-${VAGRANT_SSH} "${UNINSTALL_PKG_CMD}"
	vagrant scp "${DESTDIR}${PKGFILE}" ":${PKGFILE}"
	${VAGRANT_SSH} "${INSTALL_PKG_CMD}"
	
	# copy test scripts to VM
	vagrant scp platform_tests.tar :platform_tests.tar
	
	# run test scripts
	${VAGRANT_SSH} "tar xf platform_tests.tar"
	${VAGRANT_SSH} "sudo robot platform_tests/test.robot ||:" > testlog.txt
	
	# retrieve test reports
	mkdir -p "${DESTDIR}test_report"
	vagrant scp :log.html "${DESTDIR}test_report/"
	vagrant scp :report.html "${DESTDIR}test_report/"
	
	# finish up
	@cat testlog.txt
	@egrep '.* critical tests,.* 0 failed[[:space:]]*$$' testlog.txt\
		>/dev/null

${WORK_DIR}/${SRC_TARFILE} :
	make -C "${SRC_ROOT}" dist "DESTDIR=${WORK_DIR}/"

platform_tests.tar : $(wildcard ${SRC_ROOT}/platform_tests/**)
	tar -C "${SRC_ROOT}" -cf "$@" platform_tests

.PHONY : clean
clean :
	rm -rf "${WORK_DIR}" "${DESTDIR}${PKGFILE}" docker/src.tgz \
		testlog.txt "${DESTDIR}test_report" platform_tests.tar \
		.vm-is-running .vm-is-pristine
	-vagrant suspend

.PHONY : deepclean
deepclean : clean
	-vagrant destroy -f