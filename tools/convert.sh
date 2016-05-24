#!/bin/bash -e

find . -type l -name "*.avi" -o -name "*.mp4" -o -name "*.mkv" | while read file
do
	echo inspecting $file
	new_file=$(echo $file | sed -e 's/\(mp4\|mkv\|avi\)/webm/')
	if [ ! -e "$new_file" ]; then
		echo converting to $new_file
		echo first pass
		ffmpeg -loglevel error -hide_banner -stats -nostdin \
			-i "$file" \
			-c:v libvpx-vp9 -b:v 1400K -crf 23 -threads 3 -speed 4 -tile-columns 6 -frame-parallel 1 \
			-an \
			-pass 1 \
			-f webm \
			-y /dev/null
		echo second pass
		ffmpeg -loglevel error -hide_banner -stats -nostdin \
			-i "$file" \
			-c:v libvpx-vp9 -b:v 1400K -crf 23 -threads 3 -speed 2 \
			-tile-columns 6 -frame-parallel 1 -auto-alt-ref 1 -lag-in-frames 25 \
			-c:a libopus -b:v 64k \
			-pass 2 \
			-f webm \
			-y "$new_file"
		echo done

	else
		echo already converted
	fi
done
