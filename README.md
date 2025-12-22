# Compilation
go build -O jellyfin-latest

# Installation
```
chmod +x jellyfin-latest
sudo mv jellyfin-latest /usr/local/bin/jellyfin-latest
sudo useradd --system jellyfin-latest
```

## Setup systemd service file
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
