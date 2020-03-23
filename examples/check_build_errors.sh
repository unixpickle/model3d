#!/bin/bash

if [ $# -eq 0 ]; then
    echo 'Testing all examples...'
    find . -name 'main.go' -exec ./check_build_errors.sh {} \;
    exit
fi

cd $(dirname $1)
if ! go test 2>/dev/null >/dev/null; then
    echo "Test failure for $(dirname $1)" >&2
fi
cd - >/dev/null
