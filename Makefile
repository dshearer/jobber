DESTDIR = /usr/local
CLIENT = jobber
DAEMON = jobberd
CLIENT_USER = jobber_client

CLIENT_PERMS = 4755
DAEMON_PERMS = 0755

.PHONY : build
build :
	go install github.com/dshearer/jobber
	go install github.com/dshearer/jobber/${CLIENT}
	go install github.com/dshearer/jobber/${DAEMON}

.PHONY : install
install : build ${DESTDIR}/bin/${CLIENT} ${DESTDIR}/sbin/${DAEMON} /etc/init.d/jobber

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

.PHONY : uninstall
uninstall :
	service jobber stop
	chkconfig jobber off
	chkconfig --del jobber
	rm -f ${DESTDIR}/bin/${CLIENT} ${DESTDIR}/sbin/${DAEMON} /etc/init.d/jobber
	-userdel ${CLIENT_USER}

.PHONY : clean
clean :
	rm -rf ${GOPATH}/bin/* ${GOPATH}/pkg/*
