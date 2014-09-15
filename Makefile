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
install : build
	-userdel ${CLIENT_USER}
	useradd --home / -M --system --shell /sbin/nologin ${CLIENT_USER}
	install -d ${DESTDIR}/bin ${DESTDIR}/sbin
	install -o ${CLIENT_USER} -g root -m ${CLIENT_PERMS} -p ${GOPATH}/bin/${CLIENT} ${DESTDIR}/bin
	install -o root -g root -m ${DAEMON_PERMS} -p ${GOPATH}/bin/${DAEMON} ${DESTDIR}/sbin

.PHONY : uninstall
uninstall :
	rm -f ${DESTDIR}/bin/${CLIENT} ${DESTDIR}/sbin/${DAEMON}
	userdel ${CLIENT_USER}

.PHONY : clean
clean :
	rm -rf ${GOPATH}/bin/* ${GOPATH}/pkg/*
