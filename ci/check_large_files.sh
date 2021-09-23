#!/bin/bash

find . | while read line; do
    # Exclude list (is there a better way to write this?)
    if [[ "$line" == ./.git* ]]; then
        continue
    fi
    if [[ "$line" == ./examples/renderings/showcase/models/* ]]; then
        continue
    fi
    found="false"
    for path in ./model3d/test_data/hierarchy_test.stl.gz ./model3d/test_data/non_intersecting_hook.stl ./examples/decoration/lockbagel/printed_lockbagel.jpg ./examples/renderings/yt_banner/assets/corgi.zip ./examples/renderings/tiffany/output_hd.png; do
        if [ "$line" == "$path" ]; then
            found="true"
        fi
    done
    if [ "$found" == "true" ]; then
        continue
    fi

    if [ -f "$line" ]; then
        size=$(du -k "$line" | cut -f 1 -d $'\t' | cut -f 1 -d ' ')
        if [ "$size" -ge 1000 ]; then
            echo "file is too large: $line ($size KB)" >&2
            exit 1
        fi
    fi
done
