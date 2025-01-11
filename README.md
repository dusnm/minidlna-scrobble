# minidlna-scrobble

[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)
[![Go Report Card](https://goreportcard.com/badge/github.com/dusnm/minidlna-scrobble)](https://goreportcard.com/report/github.com/dusnm/minidlna-scrobble)

## A last.fm scrobbler for minidlna (ReadyMedia)

### Overview
This application will watch the minidlna log file for writes and scrobble plays to last.fm.

For this to work, you must set the log level of minidlna to `info` or below in its config file.
The relevant line in `/etc/minidlna` is:
```
log_level=general,artwork,database,inotify,scanner,metadata,http,ssdp,tivo=info
```

### Instalaltion
You can use the precompiled binaries for your cpu architecture (currently `amd64` and `arm64`) 
in the [release](https://github.com/dusnm/minidlna-scrobble/releases/latest) section. Make sure to put the 
binary somewhere in your `$PATH`.

An Arch Linux package is available in the AUR. You can use `makepkg` or your favorite AUR helper.
If using `makepkg` import my public key first, as it's needed to verify the package signature.
```shell
gpg --recv-keys --keyserver=hkps://keys.openpgp.org 31086781B8FA9BA0EBDA9914C303EE480C188527
```
#### Example
* makepkg
```shell
git clone https://aur.archlinux.org/minidlna-scrobble.git && cd minidlna-scrobble
makepkg -si
```

* yay
```shell
# yay should prompt you to import the GPG key, so
# manualy importing it isn't actually required.

yay -S minidlna-scrobble
```

### Building from source
To build the application, run go build with the following options:
```shell
mkdir -p ./bin && \
go build -ldflags='-s -w -extldflags "-static"' -o ./bin/minidlna-scrobble ./main.go
```

If you happen to have `go-task` installed, a `Taskfile.yml` file is included to automate the process. Just run:
```shell
task build
```
or
```shell
go-task build
```

### Logging and log level
Everything is logged to `stderr`, which you can easily redirect to any other file of your liking.
```shell
minidlna-scrobble [command] 2>>app.log
```

Every command can be assigned a log level (the default is `error`).
The levels are arranged in a hierarchical structure. The application will log all events above and including the chosen level.

The possible levels, in ascending order, are: `trace`, `debug`, `info`, `warn`, `error`, `fatal`, `panic`.

To choose a level other than the default, pass it through a command line flag:
```shell
minidlna-scrobble --log-level=info [command]
```
or with a shorthand flag
```shell
minidlna-scrobble -l info [command]
```

### Environment setup
Set your `XDG_CONFIG_HOME` and `XDG_CACHE_HOME` environment variables to writable locations.
```shell
export XDG_CONFIG_HOME="$HOME/.config"
export XDG_CACHE_HOME="$HOME/.cache"
```

You must give the application read access to the minidlna database, usually stored at `/var/cache/minidlna/files.db`,
by adding the application user to the `minidlna` group.
```shell
sudo gpasswd -a username minidlna
```

### Authenticating with last.fm
1. Apply for an API account, [here](https://www.last.fm/api/account/create) (Name and description are the only required fields)
2. You'll receive an API key and a shared secret, take note of them.
3. Write app configuration at `$XDG_CONFIG_HOME/minidlna-scrobbler/config.json`
```json
{
  "db_file": "/var/cache/minidlna/files.db",
  "log_file": "/var/log/minidlna/minidlna.log",
  "credentials": {
    "api_key": "provided_api_key",
    "shared_secret": "provided_shared_secret"
  }
}
```
4. Run the application `auth` command, and follow instructions to authorize your last.fm session
```shell
minidlna-scrobble auth
```

### Scrobbling
Run the application with the `scrobble` command to start scrobbling, there are multiple ways to do this
but using systemd is the recommended approach. Here's an example service file that you can modify to your
liking or use as-is:
```
[Unit]
Description=Scrobble to last.fm from minidlna log files
After=network.target

[Service]
Type=simple
User=username
Group=username
WorkingDirectory=/usr/local/bin
ExecStart=/usr/local/bin/minidlna-scrobble --log-level=info scrobble
Restart=on-failure

# These must be writable by the "username" user
Environment=XDG_CONFIG_HOME=/home/username/.config
Environment=XDG_CACHE_HOME=/home/username/.cache

[Install]
WantedBy=multi-user.target
```

### Notes
* The application requires go >= 1.23 to compile.
* The application assumes Linux is the underlying operating system and is therefore not portable.
There are no guarantees it'll work on other Unix-like operating systems,
and there are no guarantees whatsoever for Windows.

### Licensing
This application is free software, licensed under the terms of the GNU GPL v3 license.
