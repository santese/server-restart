package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	ping "github.com/go-ping/ping"
)

type config struct {
	serverIP         string
	pingInterval     time.Duration
	timeoutThreshold time.Duration
	idracIP          string
	idracUsername    string
	idracPassword    string
}

func main() {
	// Setup Config
	config := config{}

	flag.StringVar(&config.idracIP, "idracIP", "", "iDRAC IP address")
	flag.StringVar(&config.idracUsername, "idracUsername", "", "iDRAC username")
	flag.StringVar(&config.idracPassword, "idracPassword", "", "iDRAC password")
	flag.StringVar(&config.serverIP, "serverIP", "", "Server IP address")
	flag.DurationVar(&config.pingInterval, "pingInterval", 1*time.Minute, "Ping interval")
	flag.DurationVar(&config.timeoutThreshold, "timeoutThreshold", 5*time.Minute, "Timeout threshold")
	flag.Parse()

	var offlineDuration time.Duration

	for {
		pinger, err := ping.NewPinger(config.serverIP)
		if err != nil {
			log.Fatalf("ERROR: %s\n", err)
		}

		pinger.Count = 1

		// If we receive a response, reset the offline duration
		pinger.OnRecv = func(pkt *ping.Packet) {
			offlineDuration = 0 // Reset offline duration if we receive a response
			fmt.Printf("Received packet from %s: %d bytes\n", pkt.IPAddr, pkt.Nbytes)
		}

		// If we don't receive a response, increment the offline duration
		pinger.OnFinish = func(stats *ping.Statistics) {
			if stats.PacketLoss == 100 {
				fmt.Printf("Server has been offline for %s\n", offlineDuration)
				offlineDuration += config.pingInterval
				if offlineDuration >= config.timeoutThreshold {
					fmt.Println("Server has been offline for more than 5 minutes. Restarting...")
					restartServer(config.idracIP, config.idracUsername, config.idracPassword)
					offlineDuration = 0 // Resetting the counter
				}
			}
		}

		err = pinger.Run()
		if err != nil {
			log.Fatalf("ERROR: %s\n", err)
		}

		time.Sleep(config.pingInterval)
	}
}

func restartServer(ip, u, p string) error {
	fmt.Printf("Restarting server at %s\n", time.Now())

	// Create a new HTTP client with a 30 second timeout
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	req, err := http.NewRequest("PUT", fmt.Sprintf("https://%s/sysmgmt/2015/server/power?action=ForceRestart", ip), nil)
	if err != nil {
		return err
	}

	//TODO: Get this working, need to probably login to iDRAC first and get session cookie & potentially xsrf token
	req.SetBasicAuth(u, p)

	res, err := httpClient.Do(req)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		return err
	}

	fmt.Printf("Server restart request sent. Response: %s\n", res.Status)

	return nil
}
