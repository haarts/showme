#!/bin/bash

#path=$(printf %q "$1")
find "$1" -name '*.srt' | while read line
do
    echo "----"
    echo $line
    w=$(echo $line | sed 's/.srt//')
    if [ -f "$w.vtt" ]; then
      echo Already converted
    else
      submarine "$line"
    fi
done
