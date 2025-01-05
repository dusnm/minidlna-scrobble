# minidlna-scrobble
## A last.fm scrobbler for minidlna (ReadyMedia)

### Overview
This application will watch the minidlna log file for writes and scrobble plays to last.fm.

For this to work, you must set the log level of minidlna to `info` or below in its config file.
The relevant line in `/etc/minidlna` is:
```
log_level=general,artwork,database,inotify,scanner,metadata,http,ssdp,tivo=info
```

### Building
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

### Environment setup
Set your `XDG_CONFIG_HOME` and `XDG_CACHE_HOME` environment variables to writable locations.
```shell
export XDG_CONFIG_HOME="$HOME/.config"
export XDG_CACHE_HOME="$HOME/.cache"
```

### Authenticating with last.fm
1. Apply for an API account, [here](https://www.last.fm/api/account/create) (Name and description are the only required fields)
2. You'll receive an API key and a shared secret, take not of them.
3. Write app configuration at `$XDG_CONFIG_HOME/minidlna-scrobbler/config.json
```json
{
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
User=your_username
Group=your_username
WorkingDirectory=/usr/local/bin
ExecStart=/usr/local/bin/minidlna-scrobble scrobble
Restart=on-failure

Environment=XDG_CONFIG_HOME=/home/your_username/.config
Environment=XDG_CACHE_HOME=/home/your_username/.cache

[Install]
WantedBy=multi-user.target
```

### Notes
* The application requires go >= 1.23 to compile.
* The application assumes Linux is the underlying operating system. There are no guarantees
it'll work on other Unix-like operating systems, and that's especially true for Windows.

### Licensing
This application is free software, licensed under the terms of the GNU GPL v3 license.
