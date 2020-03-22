include packaging/alpine/sources.mk
include packaging/rpm/sources.mk
include packaging/debian/sources.mk
include packaging/darwin/sources.mk

PACKAGING_SOURCES := \
	${ALPINE_SOURCES} \
	${RPM_SOURCES} \
	${DARWIN_SOURCES} \
	${DEBIAN_SOURCES} \
	packaging/head.mk \
	packaging/Makefile \
	packaging/sources.mk \
	packaging/tail.mk
