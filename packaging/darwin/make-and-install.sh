#!/bin/bash -ex

if [ $# -ne 2 ]; then
  echo "usage: make-pkg.sh SRC_TARBALL DARWIN_DIR" >&2
  exit 1
fi

SRC_TARBALL="$1"
DARWIN_DIR="$2"

# extract source
TMPDIR="$(mktemp -d)"
WORKSPACE="${TMPDIR}/src/github.com/dshearer"
mkdir -p "${WORKSPACE}"
cd "${WORKSPACE}"
tar -xzf "${SRC_TARBALL}"

# build & install
cd jobber
mkdir pkg
make check
sudo make install "DESTDIR=/"

# run
sudo cp "${DARWIN_DIR}/launchd.plist" "${TMPDIR}/launchd.plist"
sudo launchctl unload "${TMPDIR}/launchd.plist"
sudo launchctl load "${TMPDIR}/launchd.plist"
sudo launchctl start info.nekonya.jobber

# test
sudo robot --pythonpath "${WORKSPACE}/jobber/platform_tests/keywords" \
  "${WORKSPACE}/jobber/platform_tests/suites" | tee "${DARWIN_DIR}/testlog.txt"

# clean up
sudo launchctl unload "${TMPDIR}/launchd.plist"
sudo rm -rf "${TMPDIR}"
