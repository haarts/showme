function convert_srt_to_vtt
  find $argv -name '*.srt' | while read line
    echo $line;
    set w (echo $line | sed 's/.srt//');
    if test -f "$w.vtt"
      echo Already converted
    else
      submarine $line
    end
  end
end

