package flow

import (
	"os"
	"sync"
	"time"

	"github.com/massix/protrans/pkg/nat"
	"github.com/massix/protrans/pkg/transmission"
	"github.com/sirupsen/logrus"
)

func FetchExternalIP(natClient nat.NatClientI, wg *sync.WaitGroup, ipChan chan<- string, done chan os.Signal) {
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

func MapPorts(natClient nat.NatClientI, portLifetime int, wg *sync.WaitGroup, ipChan <-chan string, portChan chan<- int, done chan os.Signal) {
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

func TransmissionArgSetter(transmissionClient transmission.TransmissionClient, wg *sync.WaitGroup, portChan <-chan int, done chan os.Signal) {
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
