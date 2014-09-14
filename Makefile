DESTDIR = /usr/local

.PHONY : build
build :
	go install github.com/dshearer/jobber
	go install github.com/dshearer/jobber/client
	go install github.com/dshearer/jobber/daemon

.PHONY : clean
clean :
	rm -rf ${GOPATH}/bin/* ${GOPATH}/pkg/*
