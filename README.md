# Statuspage

![](https://user-images.githubusercontent.com/1935975/224830896-65b21d01-62e1-4ed5-8b0c-edf717a85858.png)

I created this project because all existing self-hosted alternatives had way too many dependencies and were annoying to use. With this one you just get one binary and run it.

## Installation

### Using Go toolchain

    $ go install github.com/boreq/statuspage-backend/cmd/statuspage@latest

## Configuration

### Creating a config file

    $ statuspage default_config > config.json
    
If you want this thing to be exposed publicly without a reverse proxy server then change the address to the format `:port` instead of `1.2.3.4:port`. You need to set paths to two directories: `scriptsDirectory` and `dataDirectory`. `scriptsDirectory` contains your monitor definitions (see next section) and `dataDirectory` is where the database used by the program will be stored.

### Creating monitors

Each monitor consists of two files. The first file is a JSON file which specifies some metadata around the monitor and the second file is the script executed to check if the monitored resource is "up" or "down". Those files are named `<somename>.json` and `somename` respectfully. Make sure that the script file is executable by the user running this software. Here is an example:

    $ ls scripts
    web web.json
  
    $ cat web.json
    {
      "name": "Web server for example.com"
    }
    
    $ cat web
    #!/bin/bash
    set -e
    URL="https://example.com/"
    curl --fail -I -v -- "$URL" 2>&1

### Recommended directory structure

I usually arrange my files like this:



```
$ tree
.
├── config.json
├── data
│   ├── 000000.vlog
│   ├── LOCK
│   └── MANIFEST
└── scripts
    ├── web
    ├── web.json
    ├── api
    └── api.json
```
## Running the software

    $ statuspage run config.json
