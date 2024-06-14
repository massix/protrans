package main

import (
	"context"
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

var Version string

func main() {
	var configurationPath string

	var wg sync.WaitGroup
	defer wg.Wait()

	if len(os.Args) > 1 {
		configurationPath = os.Args[1]
		logrus.Infof("Parsing configuration from '%s'", configurationPath)
	} else {
		logrus.Info("Using default values")
	}

	conf := config.NewConfiguration(configurationPath, true)
	logrus.SetLevel(conf.LogrusLogLevel())

	logrus.Infof("Protrans version: %s", Version)
	logrus.Infof("Log level: %s", conf.LogrusLogLevel().String())
	logrus.Infof("NAT Configuration:\n%s", dumpNatConfiguration(conf))
	logrus.Infof("Transmission Configuration:\n%s", dumpTransmissionConfiguration(conf))

	natClient := nat.New(natpmp.NewClientWithTimeout(conf.GatewayIP(), 2*time.Second))
	realTransmission, err := transmissionrpc.New(conf.Transmission.Host, conf.Transmission.Username, conf.Transmission.Password, &transmissionrpc.AdvancedConfig{
		Port: conf.Transmission.Port,
	})
	if err != nil {
		logrus.Panic(err)
	}

	transmissionClient := transmission.New(realTransmission)

	ctx, cancel := context.WithCancel(context.Background())

	// Register to some signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGABRT, syscall.SIGHUP)

	// Buffered channel to make sure we're refreshing it constantly
	ipChan := make(chan string, 30)
	portChan := make(chan int, 30)

	// Avoid leaking channels
	defer func() {
		close(ipChan)
		close(portChan)
		close(sigChan)
	}()

	wg.Add(3)

	// This goroutine will constantly check the external IP address and send it to a channel
	go flow.FetchExternalIP(ctx, natClient, &wg, ipChan)

	// This goroutine will receive the IP address and create a port mapping which will be sent to another channel
	go flow.MapPorts(ctx, natClient, int(conf.Nat.PortLifetime), &wg, ipChan, portChan)

	// This goroutine will receive the mapped port and send it to Transmission if connected
	go flow.TransmissionArgSetter(ctx, transmissionClient, &wg, portChan)

	select {
	case <-ctx.Done():
		logrus.Infof("Context closed, leaving")
	case s := <-sigChan:
		logrus.Infof("Received signal %q, leaving", s)
		cancel()
	}

	logrus.Info("Waiting for goroutines to finish")
}
