#!/bin/bash

if [ $# -eq 0 ]; then
    echo 'Testing all examples...'
    find . -name 'main.go' -exec ./check_build_errors.sh {} \;
    exit
fi

if echo "$1" | grep _deprecated >/dev/null; then
    exit
fi
echo $1 | grep _deprecated

cd $(dirname $1)
if ! [ -f README.md ]; then
    echo "Missing README for $(dirname $1)" >&2
fi
if ! go test 2>/dev/null >/dev/null; then
    echo "Test failure for $(dirname $1)" >&2
fi
cd - >/dev/null
