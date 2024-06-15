package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hekmon/transmissionrpc"
	natpmp "github.com/jackpal/go-nat-pmp"
	"github.com/massix/protrans/internal/config"
	"github.com/massix/protrans/internal/nat"
	"github.com/massix/protrans/internal/transmission"
	"github.com/sirupsen/logrus"
)

var Version string

func main() {
	var configurationPath string

	if len(os.Args) > 1 {
		configurationPath = os.Args[1]
		logrus.Infof("Parsing configuration from '%s'", configurationPath)
	} else {
		logrus.Info("Using default values")
	}

	conf := config.New(configurationPath, true)
	logrus.SetLevel(conf.LogrusLogLevel())

	logrus.Infof("Protrans version: %s", Version)
	logrus.Info(conf)

	natClient := nat.New(natpmp.NewClientWithTimeout(conf.GatewayIP(), 2*time.Second))
	realTransmission, err := transmissionrpc.New(conf.Transmission.Host, conf.Transmission.Username, conf.Transmission.Password, &transmissionrpc.AdvancedConfig{
		Port: conf.Transmission.Port,
	})
	if err != nil {
		logrus.Panic(err)
	}

	transmissionClient := transmission.New(realTransmission)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	defer close(signals)

	timerDuration := 10 * time.Second
	timer := time.NewTimer(timerDuration)
	defer timer.Stop()

	running := true

	for running {
		timer.Reset(timerDuration)

		select {
		case s := <-signals:
			logrus.Infof("Received signal %q, leaving gracefully", s)
			running = false
		case <-timer.C:
			logrus.Debug("Time to check")

			if transmissionClient.IsConnected() && !transmissionClient.IsPortOpen() {
				logrus.Infof("Transmission is up @ %q", conf.Transmission.Host)

				ext, err := natClient.GetExternalAddress()
				if err != nil {
					logrus.Warnf("Could not communicate with Gateway @ %q: %s", conf.Nat.Gateway, err)
					continue
				}
				logrus.Debugf("Got external IP %q", ext)

				tcpPort, err := natClient.AddPortMapping("tcp", 600)
				if err != nil {
					logrus.Errorf("Could not add TCP port mapping: %s", err)
					continue
				}
				logrus.Debugf("Mapped TCP port: %d", tcpPort)

				udpPort, err := natClient.AddPortMapping("udp", 600)
				if err != nil {
					logrus.Errorf("Could not add UDP port mapping: %s", err)
					continue
				}
				logrus.Debugf("Mapped UDP port: %d", udpPort)

				if tcpPort != udpPort {
					logrus.Errorf("Mapped ports differ in range: TCP=%d; UDP=%d", tcpPort, udpPort)
					continue
				}

				if err := transmissionClient.SetPeerPort(tcpPort); err != nil {
					logrus.Errorf("Could not set port %d in Transmission: %s", tcpPort, err)
					continue
				}

				logrus.Debug("Port set!")
				time.Sleep(5 * time.Second)

				if transmissionClient.IsPortOpen() {
					logrus.Infof("Successfully set port %d to Transmission and checked network connectivity", tcpPort)
				} else {
					logrus.Warnf("Set port %d to Transmission but was unable to check connectivity (it might take some time...)", tcpPort)
				}
			} else {
				logrus.Debug("Transmission is not connected or port is already open")
				continue
			}
		}
	}

	logrus.Info("Closing ProTrans")
}
