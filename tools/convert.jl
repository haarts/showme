# Julia script to convert non network friendly video formats into mp4.
using Dates

cmd = `find . -type l -name "*.avi" -o -name "*.mkv"`
files = split(readchomp(cmd), "\n")
for file in files
  println(string(now(), " Converting: $file"))
  newfile = string(replace(file, r"\.(avi|mkv)$",""),".mp4")
  if !isfile(newfile)
    run(`ffmpeg -v warning -y -i $file -vcodec libx264 -tune film -preset medium -profile:v high -crf 25 -acodec libfdk_aac -ab 64k $newfile`)
    println("done")
  else
    println("skipping")
  end
end

