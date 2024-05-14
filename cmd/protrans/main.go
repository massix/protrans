package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/hekmon/transmissionrpc"
	natpmp "github.com/jackpal/go-nat-pmp"
	"github.com/massix/protrans/pkg/config"
	"github.com/massix/protrans/pkg/nat"
	"github.com/massix/protrans/pkg/transmission"
	"github.com/sirupsen/logrus"
)

func dumpNatConfiguration(conf *config.ProtransConfiguration) string {
	return fmt.Sprintf("\tGateway: %s\n\tPortLifetime: %d\n", conf.Nat.Gateway, conf.Nat.PortLifetime)
}

func dumpTransmissionConfiguration(conf *config.ProtransConfiguration) string {
	return fmt.Sprintf("\tHost: %s\n\tPort: %d\n\tUsername: %s\n", conf.Transmission.Host, conf.Transmission.Port, conf.Transmission.Username)
}

func fetchExternalIP(natClient nat.NatClientI, wg *sync.WaitGroup, ipChan chan<- string, done chan os.Signal) {
	running := true

	for running {
		ip, err := nat.GetExternalIP(natClient)
		if err != nil {
			logrus.Warn(err)
		} else {
			logrus.Debugf("Retrieved IP: %s, sending to channel", ip)
			ipChan <- ip
		}

		select {
		case s := <-done:
			logrus.Info("Gracefully stopping gateway detector")
			running = false
			done <- s // Make sure everyone is leaving
		case <-time.After(30 * time.Second):
			logrus.Debug("No signals received in 30 seconds, refreshing IP...")
		}
	}

	wg.Done()
}

func mapPorts(natClient nat.NatClientI, portLifetime int, wg *sync.WaitGroup, ipChan <-chan string, portChan chan<- int, done chan os.Signal) {
	running := true

	for running {
		select {
		case ip := <-ipChan:
			var mappedTcpPort int
			var mappedUdpPort int
			var err error
			logrus.Debugf("Mapping port for external IP: %s", ip)

			mappedTcpPort, err = nat.AddPortMapping(natClient, "tcp", portLifetime)
			if err != nil {
				logrus.Errorf("Unable to create port mapping: %s", err)
				continue
			}

			mappedUdpPort, err = nat.AddPortMapping(natClient, "udp", portLifetime)
			if err != nil {
				logrus.Errorf("Unable to create port mapping: %s", err)
				continue
			}

			if mappedTcpPort != mappedUdpPort {
				logrus.Errorf("Ports differ in range: (tcp %d) and (udp %d)", mappedTcpPort, mappedUdpPort)
				continue
			}

			logrus.Debugf("Sending port %d to channel", mappedTcpPort)
			portChan <- mappedTcpPort
		case s := <-done:
			logrus.Info("Gracefully stopping port mapper")
			running = false
			done <- s
		}
	}

	wg.Done()
}

func transmissionArgSetter(transmissionClient transmission.TransmissionClient, wg *sync.WaitGroup, portChan <-chan int, done chan os.Signal) {
	running := true

	for running {
		select {
		case mappedPort := <-portChan:
			if transmission.IsConnected(transmissionClient) {
				logrus.Debug("Transmission is connected")
				if transmission.IsPortOpen(transmissionClient) {
					logrus.Debug("Port is already set in Transmission, nothing to do")
					continue
				}

				if err := transmission.SetPeerPort(transmissionClient, mappedPort); err != nil {
					logrus.Error(err)
					continue
				}

				logrus.Debug("Port set!")
				time.Sleep(3 * time.Second)

				if transmission.IsPortOpen(transmissionClient) {
					logrus.Infof("Successfully set port %d to Transmission and checked network connectivity", mappedPort)
				} else {
					logrus.Warnf("Set port %d to Transmission but was unable to check connectivity (this may be normal, it might take some time before the NAT is recognised)", mappedPort)
				}
			} else {
				logrus.Warnf("Should set port: %d but Transmission is not connected", mappedPort)
			}
		case s := <-done:
			logrus.Info("Gracefully stopping Transmission Client")
			running = false
			done <- s
		}
	}

	wg.Done()
}

func main() {
	var configurationPath string

	if len(os.Args) > 1 {
		logrus.Infof("Parsing configuration from '%s'", configurationPath)
		configurationPath = os.Args[1]
	} else {
		logrus.Info("Using default values")
	}

	conf := config.NewConfiguration(configurationPath, true)
	logrus.SetLevel(conf.LogrusLogLevel())
	logrus.Infof("Log level: %s\n", conf.LogrusLogLevel().String())
	logrus.Infof("NAT Configuration:\n%s", dumpNatConfiguration(conf))
	logrus.Infof("Transmission Configuration:\n%s", dumpTransmissionConfiguration(conf))

	natClient := natpmp.NewClientWithTimeout(conf.GatewayIP(), 2*time.Second)
	transmissionClient, err := transmissionrpc.New(conf.Transmission.Host, conf.Transmission.Username, conf.Transmission.Password, &transmissionrpc.AdvancedConfig{
		Port: conf.Transmission.Port,
	})
	if err != nil {
		logrus.Panic(err)
	}

	// Register to some signals
	done := make(chan os.Signal, 126)
	signal.Notify(done, syscall.SIGTERM)
	signal.Notify(done, syscall.SIGINT)
	signal.Notify(done, syscall.SIGABRT)
	signal.Notify(done, syscall.SIGHUP)

	// Buffered channel to make sure we're refreshing it constantly
	ipChan := make(chan string, 30)
	portChan := make(chan int, 30)

	// Avoid leaking channels
	defer func() {
		close(ipChan)
		close(portChan)
		close(done)
	}()

	var wg sync.WaitGroup
	wg.Add(3)

	// This goroutine will constantly check the external IP address and send it to a channel
	go fetchExternalIP(natClient, &wg, ipChan, done)

	// This goroutine will receive the IP address and create a port mapping which will be sent to another channel
	go mapPorts(natClient, int(conf.Nat.PortLifetime), &wg, ipChan, portChan, done)

	// This goroutine will receive the mapped port and send it to Transmission if connected
	go transmissionArgSetter(transmissionClient, &wg, portChan, done)

	wg.Wait()
}
