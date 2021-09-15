#!/bin/bash

found=0
while read -r path; do
    out=$(gofmt -l $path)
    if [ "$out" != "" ]; then
        echo "reformatted $path" >&2
        found=$((found+1))
    fi
done <<<$(find . -name '*.go')

if [ $found -ne 0 ]; then
    echo "gofmt reformatted $found files!" >&2
    exit 1
else
    echo "no files were reformatted by gofmt"
fi