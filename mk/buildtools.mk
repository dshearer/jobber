_BUILDTOOLS_GEN = ${CURDIR}/buildtools/gen
_BUILDTOOLS_WKSPC = ${_BUILDTOOLS_GEN}/gowkspc

GOYACC = ${_BUILDTOOLS_GEN}/bin/goyacc
GO_WITH_TOOLS = PATH="${_BUILDTOOLS_GEN}/bin:$${PATH}" ${GO}

${GOYACC} : buildtools/gotools/tools-release-branch.go1.8.tar.gz
	@echo MAKE GOYACC
	@mkdir -p "${_BUILDTOOLS_WKSPC}"
	@cd buildtools/gotools/ && tar -xzf tools-release-branch.go1.8.tar.gz
	@rsync -a buildtools/gotools/tools-release-branch.go1.8/ \
		"${_BUILDTOOLS_WKSPC}/src/"
	@rm -rf buildtools/gotools/tools-release-branch.go1.8
	@cd "${_BUILDTOOLS_WKSPC}/src/golang.org/x/tools/cmd/goyacc" && \
		go build -o "${GOYACC}"

.PHONY : clean-buildtools
clean-buildtools :
	rm -rf "${_BUILDTOOLS_GEN}"