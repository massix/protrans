package flow

import (
	"context"
	"sync"
	"time"

	"github.com/massix/protrans/pkg/nat"
	"github.com/massix/protrans/pkg/transmission"
	"github.com/sirupsen/logrus"
)

func FetchExternalIP(ctx context.Context, natClient nat.Client, wg *sync.WaitGroup, ipChan chan<- string) {
	running := true

	timer := time.NewTimer(30 * time.Second)
	defer timer.Stop()

	for running {
		ip, err := natClient.GetExternalAddress()
		if err != nil {
			logrus.Warn(err)
		} else {
			logrus.Debugf("Retrieved IP: %s, sending to channel", ip)
			ipChan <- ip
		}

		select {
		case <-ctx.Done():
			logrus.Info("Gracefully stopping gateway detector")
			running = false
		case <-timer.C:
			logrus.Debug("No signals received in 30 seconds, refreshing IP...")
			timer.Reset(30 * time.Second)
		}
	}

	wg.Done()
}

func MapPorts(ctx context.Context, natClient nat.Client, portLifetime int, wg *sync.WaitGroup, ipChan <-chan string, portChan chan<- int) {
	running := true

	for running {
		select {
		case ip := <-ipChan:
			var mappedTcpPort, mappedUdpPort int
			var err error
			logrus.Debugf("Mapping port for external IP: %s", ip)

			mappedTcpPort, err = natClient.AddPortMapping("tcp", portLifetime)
			if err != nil {
				logrus.Errorf("Unable to create port mapping: %s", err)
				continue
			}

			mappedUdpPort, err = natClient.AddPortMapping("udp", portLifetime)
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
		case <-ctx.Done():
			logrus.Info("Gracefully stopping port mapper")
			running = false
		}
	}

	wg.Done()
}

func TransmissionArgSetter(ctx context.Context, transmissionClient transmission.Client, wg *sync.WaitGroup, portChan <-chan int) {
	running := true

	for running {
		select {
		case mappedPort := <-portChan:
			if transmissionClient.IsConnected() {
				logrus.Debug("Transmission is connected")
				if transmissionClient.IsPortOpen() {
					logrus.Debug("Port is already set in Transmission, nothing to do")
					continue
				}

				if err := transmissionClient.SetPeerPort(mappedPort); err != nil {
					logrus.Error(err)
					continue
				}

				logrus.Debug("Port set!")
				time.Sleep(3 * time.Second)

				if transmissionClient.IsPortOpen() {
					logrus.Infof("Successfully set port %d to Transmission and checked network connectivity", mappedPort)
				} else {
					logrus.Warnf("Set port %d to Transmission but was unable to check connectivity (this may be normal, it might take some time before the NAT is recognised)", mappedPort)
				}
			} else {
				logrus.Warnf("Should set port: %d but Transmission is not connected", mappedPort)
			}
		case <-ctx.Done():
			logrus.Info("Gracefully stopping Transmission Client")
			running = false
		}
	}

	wg.Done()
}
