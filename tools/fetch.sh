#!/bin/bash -e

cd "$(dirname "$0")"

while true; do
  start=$(date --iso-8601=seconds)
  original_file=$(ssh -oStrictHostKeyChecking=no -p$2 $1 ./first.sh)
  echo Found to be converted file: $original_file
  esc=$(printf %q "$original_file")
  echo $esc
  scp -P$2 "$1:./$esc" .
  transcoded_file=$(echo $(basename "$esc") | sed -e 's/\.\(mp4\|mkv\|avi\)$/webm/')
  echo $transcoded_file

  echo First pass
  ffmpeg -loglevel error -hide_banner -stats -nostdin \
    -i "$(basename "$original_file")" \
    -c:v libvpx-vp9 -crf 33 -b:v 0 -threads 3 -speed 2 -tile-columns 6 -frame-parallel 1 \
    -an \
    -pass 1 \
    -f webm \
    -y /dev/null
  echo Second pass
  ffmpeg -loglevel error -hide_banner -stats -nostdin \
    -i "$(basename "$original_file")" \
    -c:v libvpx-vp9 -crf 33 -b:v 0 -threads 3 -speed 1 \
    -tile-columns 6 -frame-parallel 1 -auto-alt-ref 1 -lag-in-frames 25 \
    -c:a libopus -b:a 64k \
    -pass 2 \
    -f webm \
    -y "$transcoded_file"
  echo Done

  scp -P$2 "$transcoded_file" "$1:./$(dirname "$esc")/$transcoded_file"
  rm "$transcoded_file"
  rm "$(basename "$original_file")"
  stop=$(date --iso-8601=seconds)
  echo $start,$stop,$original_file >> counter-$(hostname).csv
done
