<div align="center">

# goje
![GitHub top language](https://img.shields.io/github/languages/top/nimaaskarian/goje?color=blue)

goje (/ˈɡoʊ.dʒɛ/, meaning tomato in Persian) is an event-based pomodoro server.

[Features](#features) •
[Installation](#installation) •
[Getting started](#getting-started) •
[Usage](#usage)
</div>

## Disclaimer
goje is highly in development. i'll try my best not to change the current API, but it will inevitably change.

## Features

### Webgui
webgui is intuitive and easy to use. its run by default, and opens in your browser. it supports both light and dark mode, and uses your default browser theme.
![goje dark mode](https://github.com/user-attachments/assets/a31a8e00-22b6-4b6f-87a1-4a7bc8e0851e)


#### Custom css
if you don't like the default style of goje, you can use ``--custom-css` option and pass a custom css file to it.
here's a [pywal-themed](https://github.com/nimaaskarian/goje/wiki/Pywal-integration) goje using this feature:

![pywal-themed goje using custom css](https://github.com/user-attachments/assets/d00fa5cd-ab5d-442f-a195-1b233283b896)



## ActivityWatch
You can turn on the activitywatch watcher using the config option `activitywatch` (`--activitywatch` cli argument) 
![activitywatch bucket](https://github.com/user-attachments/assets/3bd1ffc6-1cc7-4a6a-a110-728ee1823507)

## Installation

### From source
clone the repository, and run `make` to install dependencies and compile, or
`sudo make install` to compile and install. note that `go build` won't suffice.
Use `make`.

```
git clone https://github.com/nimaaskarian/goje
sudo make install
```

### Prebuilt binary
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
