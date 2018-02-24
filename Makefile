.PHONY : main
main :
	@echo "build - generate pages"
	@echo "relinfo - download release info from GitHub"

.PHONY : build
build :
	php build.php

.PHONY : relinfo
relinfo :
	python3 get-release-info.py > phplib/latest-release.json
