package main

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	ping "github.com/go-ping/ping"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
)

type Config struct {
	ServerIP         string        `yaml:"serverIP"`
	PingInterval     time.Duration `yaml:"pingInterval"`     // in minutes
	TimeoutThreshold time.Duration `yaml:"timeoutThreshold"` // in minutes
	IdracIP          string        `yaml:"idracIP"`
	IdracUsername    string        `yaml:"idracUsername"`
	IdracPassword    string        `yaml:"idracPassword"`
	RestartTimeout   time.Duration `yaml:"restartTimeout"` // in minutes
}

func LoadConfig() (config Config, err error) {
	viper.AddConfigPath(".")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	err = viper.ReadInConfig()
	if err != nil {
		return config, err
	}

	err = viper.Unmarshal(&config)
	if err != nil {
		return config, err
	}

	config.PingInterval = config.PingInterval * time.Minute
	config.TimeoutThreshold = config.TimeoutThreshold * time.Minute
	config.RestartTimeout = config.RestartTimeout * time.Minute

	return config, nil
}

func main() {
	config, err := LoadConfig()

	if err != nil {
		log.Fatalf("Error loading config: %s\n", err)
	}

	var offlineDuration time.Duration

	for {
		pinger, err := ping.NewPinger(config.ServerIP)
		if err != nil {
			log.Fatalf("Error starting ping: %s\n", err)
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
				offlineDuration += config.PingInterval
				if offlineDuration >= config.TimeoutThreshold {
					fmt.Printf("Server has been offline for more than %s minutes. Restarting...\n", config.TimeoutThreshold)
					err = restartServer(config.IdracIP, config.IdracUsername, config.IdracPassword)
					// Sleep for a bit to give the server time to restart
					time.Sleep(config.RestartTimeout * time.Minute)
					if err != nil {
						fmt.Printf("Error Restarting Server: %s\n", err)
					}
					offlineDuration = 0 // Resetting the counter
				}
			}
		}

		err = pinger.Run()
		if err != nil {
			log.Fatalf("Error running ping: %s\n", err)
		}

		time.Sleep(config.PingInterval)
	}
}

// Restarts idrac via ssh
// Credit https://github.com/jhunt/buffalab/blob/master/idrac.go
func restartServer(ip, username, password string) error {
	client, err := ssh.Dial("tcp", ip+":22", &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: func(host string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	})
	if err != nil {
		return fmt.Errorf("failed to dial: %s", err)
	}

	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	var out bytes.Buffer
	session.Stdout = &out
	if err := session.Run("racadm serveraction hardreset"); err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "reset %s via idrac/racadm (output follows):%s\n", ip, out.String())
	return nil
}
