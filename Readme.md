# Automatic Restarts for Dell idrac Server

Needed to automatically restart my unraid server if any kernal panics occurred. I was having issues with macvlan networking and wanted a way to auto restart it.

This package requires Dell idrac to be setup to trigger the restart of the server. It waits for x minutes of packet loss before triggering the reboot. 

Example config.yaml:

```
serverIp: 192.168.0.1
pingInterval: 1 # in minutes
timeoutThreshold: 5 # in minutes
idracIp: 192.168.0.2
idracUsername: user
idracPassword: password

```

