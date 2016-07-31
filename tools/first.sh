#!/bin/bash -e

find Videos -type l -name "*.avi" -o -name "*.mp4" -o -name "*.mkv" | while read file
do
	new_file=$(echo $file | sed -e 's/\(mp4\|mkv\|avi\)/webm/')
	if [ ! -e "$new_file" ]; then
		touch "$new_file"
		echo $file
		break
	fi
done
