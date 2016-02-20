#!/bin/sh

CURR_COMMIT=`git log -1 --pretty=format:%H`
UPDATE=N
if [ ! -f version.go ]; then
    UPDATE=Y
elif ! grep "${CURR_COMMIT}" version.go >/dev/null; then
    UPDATE=Y
elif [ version.go.in -nt version.go ]; then
    UPDATE=Y
fi
if [ "${UPDATE}" = "Y" ]; then
    sed "s/{{JobberRevision}}/${CURR_COMMIT}/g" version.go.in > version.go
fi
