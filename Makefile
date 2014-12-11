GOPATH ?= ${HOME}/go_workspace

DESTDIR = /usr/local
CLIENT = jobber
DAEMON = jobberd
CLIENT_USER = jobber_client

CLIENT_PERMS = 4755
DAEMON_PERMS = 0755

SE_FILES = se_policy/jobber.fc \
           se_policy/jobber.if \
           se_policy/jobber.te \
           ${wildcard se_policy/include/**} \
           se_policy/Makefile \
           se_policy/policygentool

.PHONY : build
build :
	go get code.google.com/p/go.net/context
	go get gopkg.in/yaml.v2
	go install github.com/dshearer/jobber
	go install github.com/dshearer/jobber/${CLIENT}
	go install github.com/dshearer/jobber/${DAEMON}

.PHONY : install
install : build \
          ${DESTDIR}/bin/${CLIENT} \
          ${DESTDIR}/sbin/${DAEMON} \
          /etc/init.d/jobber \
          /var/lock/subsys/jobber \
          se_policy/.installed

${DESTDIR}/bin/${CLIENT} : ${GOPATH}/bin/${CLIENT}
	-userdel ${CLIENT_USER}
	useradd --home / -M --system --shell /sbin/nologin ${CLIENT_USER}
	install -d ${DESTDIR}/bin
	install -T -o ${CLIENT_USER} -g root -m ${CLIENT_PERMS} -p ${GOPATH}/bin/${CLIENT} $@

${DESTDIR}/sbin/${DAEMON} : ${GOPATH}/bin/${DAEMON}
	install -d ${DESTDIR}/sbin
	install -T -o root -g root -m ${DAEMON_PERMS} -p ${GOPATH}/bin/${DAEMON} $@

/etc/init.d/jobber : jobber_init
	install -T -o root -g root -m 755 $< $@
	chkconfig --add jobber
	chkconfig jobber on

/var/lock/subsys/jobber : ${DESTDIR}/sbin/${DAEMON} /etc/init.d/jobber
	service jobber restart

se_policy/jobber.pp : ${SE_FILES}
	${MAKE} -C se_policy

se_policy/.installed : se_policy/jobber.pp
	semodule -i $< -v
	restorecon -Rv /usr/local /etc/init.d
	touch $@

.PHONY : uninstall
uninstall :
	service jobber stop
	chkconfig jobber off
	chkconfig --del jobber
	rm -f ${DESTDIR}/bin/${CLIENT} ${DESTDIR}/sbin/${DAEMON} /etc/init.d/jobber
	-userdel ${CLIENT_USER}
	semodule -r jobber -v
	rm -f se_policy/.installed

.PHONY : clean
clean :
	go clean -i github.com/dshearer/jobber
	go clean -i github.com/dshearer/jobber/${CLIENT}
	go clean -i github.com/dshearer/jobber/${DAEMON}
	${MAKE} -C se_policy clean
