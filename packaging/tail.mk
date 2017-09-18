.PHONY : main
main :
	@echo "Choose pkg-local or pkg-vm or test-vm"

.PHONY : pkg-vm
pkg-vm : ${DESTDIR}${PKGFILE}

${DESTDIR}${PKGFILE} : Vagrantfile ${WORK_DIR}/${SRC_TARFILE} \
		${PKGFILE_DEPS}
	# restore "Base" snapshot and start VM
	(vagrant snapshot list | grep Base >/dev/null) || \
		(vagrant up && vagrant reload && vagrant suspend && \
			vagrant snapshot save Base)
	vagrant snapshot restore Base
	
	# copy Jobber source to VM
	vagrant scp "${WORK_DIR}/${SRC_TARFILE}" ":${SRC_TARFILE}"
	vagrant ssh -c "tar -xzmf ${SRC_TARFILE}"
	
	# make Jobber package
	vagrant ssh -c "make -C \
		jobber-${VERSION}/packaging/${PACKAGING_SUBDIR} pkg-local \
		DESTDIR=~/"
	
	# copy package out of VM
	vagrant scp :${PKGFILE_VM_PATH} "${DESTDIR}${PKGFILE}"
	
	# stop VM
	vagrant suspend
	
	touch "$@"

.PHONY : test-vm
test-vm : ${DESTDIR}${PKGFILE} robot_tests.tar
	# restore "Base" snapshot and start VM
	vagrant snapshot restore Base
	
	# install package
	vagrant scp "${DESTDIR}${PKGFILE}" ":${PKGFILE}"
	vagrant ssh -c "${INSTALL_PKG_CMD}"
	
	# copy test scripts to VM
	vagrant scp robot_tests.tar :robot_tests.tar
	
	# run test scripts
	vagrant ssh -c "tar xf robot_tests.tar"
	vagrant ssh -c "sudo robot robot_tests/test.robot ||:" > testlog.txt
	
	# retrieve test reports
	mkdir -p "${DESTDIR}test_report"
	vagrant scp :log.html "${DESTDIR}test_report/"
	vagrant scp :report.html "${DESTDIR}test_report/"
	
	# stop VM
	vagrant suspend
	
	# finish up
	@cat testlog.txt
	@egrep '.* critical tests,.* 0 failed[[:space:]]*$$' testlog.txt\
		>/dev/null

${WORK_DIR}/${SRC_TARFILE} :
	make -C "${SRC_ROOT}" dist "DESTDIR=${WORK_DIR}/"

robot_tests.tar : ${ROBOT_TESTS}
	tar -C ${ROBOT_TESTS_DIR}/.. -cf "$@" robot_tests

.PHONY : clean
clean :
	rm -rf "${WORK_DIR}" "${DESTDIR}${PKGFILE}" docker/src.tgz \
		testlog.txt "${DESTDIR}test_report" robot_tests.tar
	-vagrant suspend

.PHONY : deepclean
deepclean : clean
	vagrant destroy -f