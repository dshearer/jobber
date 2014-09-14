DESTDIR = /usr/local

.PHONY : build
build :
	go install github.com/dshearer/jobber
	go install github.com/dshearer/jobber/client
	go install github.com/dshearer/jobber/daemon

.PHONY : install
install : build
	cp ${GOPATH}/bin/client ${DESTDIR}/bin/client
	cp ${GOPATH}/bin/daemon ${DESTDIR}/bin/daemon

.PHONY : clean
clean :
	rm -rf ${GOPATH}/bin/* ${GOPATH}/pkg/*
