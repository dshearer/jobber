include packaging/centos_6/sources.mk
include packaging/centos_7/sources.mk
include packaging/debian_9/sources.mk
include packaging/ubuntu_16/sources.mk

PACKAGING_SOURCES := \
	${CENTOS_6_SOURCES} \
	${CENTOS_7_SOURCES} \
	${DEBIAN_9_SOURCES} \
	${UBUNTU_16_SOURCES} \
	packaging/head.mk \
	packaging/Makefile \
	packaging/sources.mk \
	packaging/tail.mk