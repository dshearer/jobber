.PHONY : build
build :
	php build.php

.PHONY : relinfo
relinfo :
	python get-release-info.py > phplib/latest-release.json