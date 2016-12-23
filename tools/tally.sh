#!/bin/bash

find Public -maxdepth 1 -type d -print0 | while read -d $'\0' dir
do
        if [ "$dir" != "Public" ]; then
                converted=$(find "$dir" -name "*webm" | wc -l)
                unconverted=$(find "$dir" -type l -name '*avi' -o -name '*mp4' -o -name '*mkv' | wc -l)

                echo "$dir"
                echo "unconverted: $unconverted"
                echo "converted:   $converted"
        fi
done

