# Server Restart

## Work in Progress

Needed to automatically restart my unraid server if any kernal panics occur. I was having issues with macvlan networking  and wanted to way to auto restart if i wasn't home. 

This pacakge requires Dell Idrac to be setup to trigger the restart of the server. It waits for x minutes of packet loss before triggering the reboot. 

Flags

```
	serverIP         string
	pingInterval     time.Duration
	timeoutThreshold time.Duration
	idracIP          string
	idracUsername    string
	idracPassword    string
```

