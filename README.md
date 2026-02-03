<div align="center">

 <img src="https://github.com/nimaaskarian/goje/blob/master/httpd/webgui-preact/public/assets/goje-192x192.png" width="90" height="90" title="goje"/>
 
# goje


![GitHub top language](https://img.shields.io/github/languages/top/nimaaskarian/goje?color=blue)

goje (/ˈɡoʊ.dʒɛ/, meaning tomato in Persian) is an event-based pomodoro server;
a fast, multi client and featureful pomodoro app that uses low resources (~6MB
of RAM according to systemd's status, while using all the features and a couple
of clients) and just dissolves into your setup, whatever that might be!

#### goje on phone, goje on bar, goje on browser, goje everywhere
![goje on phone, goje on bar, goje on chrome, goje everywhere](https://github.com/user-attachments/assets/da2bf498-802b-43ba-8ed9-e462f7b0d0bf)

[Installation](#installation) •
[Getting started](#getting-started) •
[Features](#features) •
[Integration and customization](#integration-and-customization)
</div>

## Installation

### Source
clone the repository, and run `make` to install dependencies and compile, or
`sudo make install` to compile and install. note that `go build` won't suffice.
Use `make`.

```
git clone https://github.com/nimaaskarian/goje
sudo make install
```

### Binary
to use prebuilt binaries, checkout
[releases](https://github.com/nimaaskarian/goje/releases). Windows binary is
shipped with a `goje-launcher.bat` that allows you to run goje with a single
click (if the Windows security thing doesn't scream at you, because the app is
not signed by a developer key).

#### *nix
the *nix versions are bzipped; after downloading, use `bunzip2` to unzip
them:

```bash
# on linux
bunzip2 goje_linux_amd64.bz2
chmod +x goje_linux_amd64
# on termux
bunzip2 goje_android_arm64.bz2
chmod +x goje_android_arm64
```

## Getting started
after [installation](#installation), you can run goje on command line using
`goje`; or if you're on Windows, you can download the [launcher batch
script](https://github.com/nimaaskarian/goje/blob/master/goje-launcher.bat) to
run goje with a single click. this launcher already comes with the Windows
version in [releases](https://github.com/nimaaskarian/goje/releases).

## Features

### ActivityWatch
You can turn on the activitywatch watcher using the config option `activitywatch = true` (`--activitywatch` cli argument) 
![activitywatch bucket](https://github.com/user-attachments/assets/3bd1ffc6-1cc7-4a6a-a110-728ee1823507)

### Webgui
webgui is intuitive and easy to use. its run by default, and opens in your browser. it supports both light and dark modes, based on your browser's default.
![goje dark mode](https://github.com/user-attachments/assets/a31a8e00-22b6-4b6f-87a1-4a7bc8e0851e)


### Custom css
if you don't like the default style of goje webgui, you can use the config option `custom-css = /path/to/style.css`(`--custom-css /path/to/style.css` cli argument)
option and pass a custom css file to it. here's a
[pywal-themed](https://github.com/nimaaskarian/goje/wiki/Pywal-integration) goje
using this feature:

![pywal-themed goje using custom css](https://github.com/user-attachments/assets/d00fa5cd-ab5d-442f-a195-1b233283b896)

### MPRIS
you can use `mpris = true` config option (`--mpris` cli argument) to expose a
MPRIS interface on your session dbus. this means you can control goje using a
MPRIS interface, from your gnome bar, using playerctl, or using bluetooth
devices and mpris-proxy. 

thanks to [mpd-mpris](https://github.com/natsukagami/mpd-mpris) this feature. i
shamelessly copied most of their code.


### Ntfy
you can use `ntfy-address = http://some.ntfy.server/some-topic` (`--ntfy-address
http://some.ntfy.server/some-topic` cli argument) to send notifications directly
to a ntfy server.

you can also use `ntfy-auth = username:password` (`--ntfy-auth
username:password` cli argument) to specify username and password for your
password protected topic.


### Client (inbound server mirroring outbound server)
goje using its client subcommand `goje client` can mirror an outbound server
and run an inbound server. this way any action done to any of the servers are
reflected in all clients (including other inbound servers using either of these
servers as an outbound).

a `goje client` can get all options implemented for `goje` server itself. in
addition to a required `--outbound-address https://some.goje-server.org` which
specifies the outbound goje's http(s) server. the server maybe proxied behind
nginx or some sort of a webserver (`https://some.server.org/goje` for example)

## Integration and customization
checkout [wiki](https://github.com/nimaaskarian/goje/wiki) for more indepth
configuration options.
