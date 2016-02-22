GOPATH ?= ${HOME}/go_workspace

DESTDIR = /usr/local
CLIENT = jobber
DAEMON = jobberd
CLIENT_USER = jobber_client
TEST_TMPDIR = ${PWD}

SE_FILES = se_policy/jobber.fc \
           se_policy/jobber.if \
           se_policy/jobber.te \
           ${wildcard se_policy/include/**} \
           se_policy/Makefile \
           se_policy/policygentool

.PHONY : default
default : build test

.PHONY : build
build : version.go
	go install github.com/dshearer/jobber
	go install github.com/dshearer/jobber/${CLIENT}
	go install github.com/dshearer/jobber/${DAEMON}

.PHONY : test
test :
	TMPDIR="${TEST_TMPDIR}" go test github.com/dshearer/jobber/jobberd

.PHONY : install-bin
install-bin : build \
              test \
              ${DESTDIR}/bin/${CLIENT} \
              ${DESTDIR}/sbin/${DAEMON}

.PHONY : install-centos
install-centos : build \
                 test \
                 /etc/init.d/jobber \
                 /var/lock/subsys/jobber \
                 se_policy/.installed

.PHONY : install
install : build install-bin install-centos

${DESTDIR}/bin/${CLIENT} : ${GOPATH}/bin/${CLIENT}
	-userdel ${CLIENT_USER}
	useradd --home / -M --system --shell /sbin/nologin "${CLIENT_USER}"
	install -d "${DESTDIR}/bin"
	install -T -o "${CLIENT_USER}" -g root -m 4755 -p "${GOPATH}/bin/${CLIENT}" "$@"

${DESTDIR}/sbin/${DAEMON} : ${GOPATH}/bin/${DAEMON}
	install -d "${DESTDIR}/sbin"
	install -T -o root -g root -m 0755 -p "${GOPATH}/bin/${DAEMON}" "$@"

/etc/init.d/jobber : jobber_init
	install -T -o root -g root -m 0755 "$<" "$@"
	chkconfig --add jobber
	chkconfig jobber on

/var/lock/subsys/jobber : ${DESTDIR}/sbin/${DAEMON} /etc/init.d/jobber
	service jobber restart

se_policy/.installed : ${SE_FILES}
	-[ -f /etc/sysconfig/selinux ] && ${MAKE} -C se_policy && semodule -i "$<" -v && restorecon -Rv /usr/local /etc/init.d
	touch "$@"

.PHONY : uninstall-bin
uninstall-bin :
	rm -f "${DESTDIR}/bin/${CLIENT}" "${DESTDIR}/sbin/${DAEMON}"
	-userdel "${CLIENT_USER}"

.PHONY : uninstall-centos
uninstall-centos :
	service jobber stop
	chkconfig jobber off
	chkconfig --del jobber
	rm -f /etc/init.d/jobber
	-[ -f /etc/sysconfig/selinux ] && semodule -r jobber -v
	rm -f se_policy/.installed

.PHONY : uninstall
uninstall : uninstall-bin uninstall-centos

.PHONY : clean
clean :
	go clean -i github.com/dshearer/jobber
	go clean -i "github.com/dshearer/jobber/${CLIENT}"
	go clean -i "github.com/dshearer/jobber/${DAEMON}"
	-${MAKE} -C se_policy clean
	rm -f version.go
