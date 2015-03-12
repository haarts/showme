# ShowMe
A media server with no options what so ever.

I don't want all these options, it will just confuse my spouse. Convention over
configuration!

This project has two executables:
- Conjoiner, this walks a directory structure containing shows and creates JSON
  files (containing pointers to images etc etc). The resulting files are be be
  used by ShowMe.
- ShowMe, a trivial file server which serves up these JSONs _and_ a Javascript
  (Vue) application.


# Required directory structure.

```
shows
|-- Name
|    |-- 1
|    |   |-- Name 1x01-Title.avi
|    |   +-- Name 1x02-Title.avi
|    |-- 2
|    |   |-- Name 1x01-Title.avi
|    |   +-- Name 1x02-Title.avi
|
|-- Other Name
     |-- 1

```
