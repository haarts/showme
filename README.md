# ShowMe
[![Build
Status](https://snap-ci.com/haarts/showme/branch/master/build_image)](https://snap-ci.com/haarts/showme/branch/master)

A browser based media server with no options what so ever.

I don't want all these options, it will just confuse my spouse.

This project contains one executable called Fetcher. Run it with:
```
$ ./fetcher your-video-root-directory
```

Based on the required directory structure and the WebM files it contains it
will:
1. create a bunch of JSON files containing show information.
1. creates a subdirectory for every episode.
1. creates, in every directory, an `index.html`.

These `index.html` files are 'apps'. There's a:
1. Shows app
1. Show app
1. Season app
1. Episode app

Next spin up your favourite webserver with the correct document root and you're
ready to watch, in your browser.

# Required directory structure.
```
shows
|-- Name
|    |-- 1
|    |   |-- Name 1x01-Title.webm
|    |   +-- Name 1x02-Title.webm
|    |-- 2
|    |   |-- Name 1x01-Title.webm
|    |   +-- Name 1x02-Title.webm
|
|-- Other Name
     |-- 1

```
