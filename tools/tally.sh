#!/bin/bash

find Public -maxdepth 1 -type d -print0 | sort -z | while read -d $'\0' dir
do
        if [ "$dir" != "Public" ]; then
                converted=$(find "$dir" -name "*webm" | wc -l)
                unconverted=$(find "$dir" -type l -name '*avi' -o -name '*mp4' -o -name '*mkv' | wc -l)

                showname=$(echo $dir | sed 's/Public\///')
                showname_without_year=$(echo $showname | sed -E 's/ \([[:digit:]]+\)//' | sed 's/ /%20/g')
                available=$(curl --silent "http://api.tvmaze.com/singlesearch/shows?q=$showname_without_year&embed=episodes" | jq '._embedded.episodes | length')
                # '%-3d' means: line out to the left
                printf "%3d/%3d/%3d : %s\n" $converted $unconverted $available "$showname"
        fi
done

