# Server Restart

## Work in Progress

Needed to automatically restart my unraid server if any kernal panics occur. I was having issues with macvlan networking  and wanted to way to auto restart if i wasn't home. 

This package requires Dell Idrac to be setup to trigger the restart of the server. It waits for x minutes of packet loss before triggering the reboot. 

Looks for config.yaml in the same directory 

Example config.yaml:

```
serverIp: 192.168.0.1
pingInterval: 1 # in minutes
timeoutThreshold: 5 # in minutes
idracIp: 192.168.0.2
idracUsername: user
idracPassword: password


```

