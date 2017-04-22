#!/bin/bash

source "$(dirname "$0")/secrets.sh"

find Public -maxdepth 1 -type d -print0 | sort -z | while read -d $'\0' dir
do
        if [ "$dir" != "Public" ]; then
                transcoded=$(find "$dir" -name "*webm" | wc -l)
                renamed=$(find "$dir" -type l -name '*avi' -o -name '*mp4' -o -name '*mkv' | wc -l)

                showname=$(echo $dir | sed 's/Public\///')

                # Convert it so TvMaze doesn't choke
                showname_without_year=$(echo $showname | sed -E 's/ \([[:digit:]]+\)//' | sed 's/ /%20/g')
                available=$(curl --silent "http://api.tvmaze.com/singlesearch/shows?q=$showname_without_year&embed=episodes" | jq '._embedded.episodes | length')

                # Convert it to something 'safe' by converting to lower case and replacing all non alpha numerics
                showname_safe=$(echo $showname | tr '[:upper:]' '[:lower:]' | tr -c '[:alnum:]' '-' | sed 's/.$//')
                curl --silent --insecure "https://$user:$password@$server:$port/write?db=telegraf" \
                  --data-binary "progress,serie=$showname_safe transcoded=$transcoded,renamed=$renamed,available=$available"
                #echo "progress,serie=$showname_safe transcoded=$transcoded,renamed=$renamed,available=$available"
        fi
done

