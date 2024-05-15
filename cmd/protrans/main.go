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
	"github.com/massix/protrans/pkg/flow"
	"github.com/sirupsen/logrus"
)

func dumpNatConfiguration(conf *config.ProtransConfiguration) string {
	return fmt.Sprintf("\tGateway: %s\n\tPortLifetime: %d\n", conf.Nat.Gateway, conf.Nat.PortLifetime)
}

func dumpTransmissionConfiguration(conf *config.ProtransConfiguration) string {
	return fmt.Sprintf("\tHost: %s\n\tPort: %d\n\tUsername: %s\n", conf.Transmission.Host, conf.Transmission.Port, conf.Transmission.Username)
}

func main() {
	var configurationPath string

	if len(os.Args) > 1 {
		configurationPath = os.Args[1]
		logrus.Infof("Parsing configuration from '%s'", configurationPath)
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
	go flow.FetchExternalIP(natClient, &wg, ipChan, done)

	// This goroutine will receive the IP address and create a port mapping which will be sent to another channel
	go flow.MapPorts(natClient, int(conf.Nat.PortLifetime), &wg, ipChan, portChan, done)

	// This goroutine will receive the mapped port and send it to Transmission if connected
	go flow.TransmissionArgSetter(transmissionClient, &wg, portChan, done)

	wg.Wait()
}
