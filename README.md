# Jellyfin Latest
A tiny golang application that provides an API and HTML widget endpoint for Jellyfin's Latest movies, music, shows, books, and more.

![Screenshot](github/screenshot.jpg)

## Compilation
`go build -O jellyfin-latest`

## Installation
```
chmod +x jellyfin-latest
sudo mv jellyfin-latest /usr/local/bin/jellyfin-latest
sudo useradd --system jellyfin-latest
```

### Configuration
Add a .env to /usr/local/bin (if not already exists) with the following variables:
```
JELLYFIN_URL = "http://192.168.1.69:8096"
JELLYFIN_TOKEN = ""
JELLYFIN_USERID = ""
```

### Setup systemd service file
`sudo nano /etc/systemd/system/jellyfin-latest.service`

with content:
```
[Unit]
Description=Jellyfin Latest Service
After=network.target

[Service]
Type=simple
# The full path to your binary
ExecStart=/usr/local/bin/jellyfin-latest
User=jellyfin-latest
# Restart the service if it fails
Restart=always
StandardOutput=syslog
StandardError=syslog
SyslogIdentifier=jellyfin-latest

[Install]
WantedBy=multi-user.target
```

afterwards you can interact via systemctl:
```
sudo systemctl enable jellyfin-latest.service
sudo systemctl start jellyfin-latest.service
```
