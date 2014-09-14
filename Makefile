DESTDIR = /usr/local

.PHONY : build
build :
	go install github.com/dshearer/jobber
	go install github.com/dshearer/jobber/client
	go install github.com/dshearer/jobber/daemon

.PHONY : install
install : build
	install -o root -g root -d ${DESTDIR}/bin
	install -o root -g root ${GOPATH}/bin/client ${DESTDIR}/bin/client ${DESTDIR}/bin

.PHONY : clean
clean :
	rm -rf ${GOPATH}/bin/* ${GOPATH}/pkg/*
