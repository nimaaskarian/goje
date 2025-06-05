<div align="center">

# goje
![GitHub top language](https://img.shields.io/github/languages/top/nimaaskarian/goje?color=blue)

goje (/ˈɡoʊ.dʒɛ/, meaning tomato in Persian) is an event-based pomodoro server.

[Features](#features) •
[Installation](#installation) •
[Getting started](#getting-started) •
[Usage](#usage)
</div>

## Features
### Webgui
webgui is intuitive and easy to use. the server listens to `localhost:7900` for
it by default.

## ActivityWatch watcher
You can turn on the activitywatch watcher using the config option `activitywatch` (`--activitywatch` cli argument) 

## Installation
### from source
clone the repository, and run `make` to install dependencies and compile, or
`sudo make install` to compile and install. note that `go build` won't suffice.
Use `make`.

```
git clone https://github.com/nimaaskarian/goje
sudo make install
```

### binary
to use prebuilt binaries, checkout [releases](https://github.com/nimaaskarian/goje/releases).

## Getting started
after [installation](#installation), you can run goje on command line using
`goje`. or if you're on windows, you can download the [launcher batch
script](https://github.com/nimaaskarian/goje/blob/master/goje-launcher.bat) to
run goje with a single click.

## Usage
if the guides here wasn't enough, or the config options are still unclear,
checkout [wiki](https://github.com/nimaaskarian/goje/wiki) for more indepth
configuration options.
