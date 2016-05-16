SRC_ARCHIVE=jobber_src.tar.blah
SRC_ROOT=../..

.PHONY : build
build : ${SRC_ARCHIVE} pkgrel
	abuild checksum
	env "pkgrel=`cat pkgrel`" "pkgver=`cat ${SRC_ROOT}/version`" abuild -r

.PHONY : ${SRC_ARCHIVE}
${SRC_ARCHIVE} :
	cd ${SRC_ROOT} && git archive --format=tar --prefix=jobber/ -o "packaging/alpine/$@" HEAD

.PHONY : clean
clean :
	rm -f "${SRC_ARCHIVE}"
	-abuild clean
	-abuild cleanoldpkg
	-abuild cleanpkg
	-abuild cleancache
