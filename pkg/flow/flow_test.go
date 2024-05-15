package flow_test

import (
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/hekmon/transmissionrpc"
	natpmp "github.com/jackpal/go-nat-pmp"
	"github.com/massix/protrans/pkg/flow"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var (
	externalIPRetrieved bool
	portMapped          bool
	peerPortSet         bool
)

// Create the fake interfaces
type fakeNatClient struct {
	CanGetExternalIP bool
	CanMapPort       bool
}

func (n *fakeNatClient) GetExternalAddress() (*natpmp.GetExternalAddressResult, error) {
	if n.CanGetExternalIP {
		externalIPRetrieved = true
		return &natpmp.GetExternalAddressResult{
			ExternalIPAddress: [4]byte{10, 20, 30, 40},
		}, nil
	}

	return nil, errors.New("NAT not connected")
}

func (n *fakeNatClient) AddPortMapping(protocol string, internalPort, requestedExternalPort, lifetime int) (*natpmp.AddPortMappingResult, error) {
	if n.CanMapPort {
		portMapped = true
		return &natpmp.AddPortMappingResult{
			MappedExternalPort: 4242,
		}, nil
	}

	return nil, errors.New("Can not map port")
}

type fakeTransmissionClient struct {
	IsConnected       bool
	IsPortAlreadyOpen bool
	CanSetArgument    bool
}

func (ft *fakeTransmissionClient) PortTest() (bool, error) {
	return ft.IsConnected && ft.IsPortAlreadyOpen, nil
}

func (ft *fakeTransmissionClient) SessionStats() (*transmissionrpc.SessionStats, error) {
	if ft.IsConnected {
		return &transmissionrpc.SessionStats{}, nil
	}

	return nil, errors.New("Not connected")
}

func (ft *fakeTransmissionClient) SessionArgumentsSet(*transmissionrpc.SessionArguments) error {
	if ft.CanSetArgument {
		peerPortSet = true
		return nil
	}

	return errors.New("Cannot set argument")
}

// We do not really care about this function
func (ft *fakeTransmissionClient) SessionArgumentsGet() (*transmissionrpc.SessionArguments, error) {
	panic("Should not call this!")
}

func reset() {
	logrus.SetLevel(logrus.DebugLevel)
	externalIPRetrieved = false
	portMapped = false
	peerPortSet = false
}

// Happy path test, everything is ok and working
func Test_Flow_OK(t *testing.T) {
	var wg sync.WaitGroup
	reset()
	wg.Add(3)

	nc := &fakeNatClient{true, true}
	tc := &fakeTransmissionClient{true, false, true}

	ipChan := make(chan string)
	portChan := make(chan int)
	done := make(chan os.Signal, 1)
	defer func() {
		close(ipChan)
		close(portChan)
		close(done)
	}()

	// Start the flow
	go flow.FetchExternalIP(nc, &wg, ipChan, done)
	go flow.MapPorts(nc, 600, &wg, ipChan, portChan, done)
	go flow.TransmissionArgSetter(tc, &wg, portChan, done)

	<-time.After(3 * time.Second)
	done <- os.Interrupt

	// Wait for everyone to finish
	wg.Wait()

	assert.True(t, peerPortSet)
	assert.True(t, externalIPRetrieved)
	assert.True(t, portMapped)
}

// We are unable to retrieve the external IP (probably not connected to the internet?)
func Test_Flow_NoExternalIP(t *testing.T) {
	var wg sync.WaitGroup
	reset()
	wg.Add(1)

	nc := &fakeNatClient{false, true}

	ipChan := make(chan string)
	done := make(chan os.Signal, 1)
	defer func() {
		close(ipChan)
		close(done)
	}()

	go flow.FetchExternalIP(nc, &wg, ipChan, done)

	// We should fall into the timeout here
	select {
	case ip := <-ipChan:
		t.Fatalf("Should not have received an IP, received: %s instead?", ip)
	case <-time.After(5 * time.Second):
		done <- os.Interrupt
	}

	wg.Wait()

	assert.False(t, peerPortSet)
	assert.False(t, externalIPRetrieved)
	assert.False(t, portMapped)
}

// We are connected to the internet but unable to open ports (probably the Gateway is offline?)
func Test_Flow_NoPortMapping(t *testing.T) {
	var wg sync.WaitGroup
	reset()
	wg.Add(2)

	nc := &fakeNatClient{true, false}

	ipChan := make(chan string)
	portChan := make(chan int)
	done := make(chan os.Signal, 1)
	defer func() {
		close(ipChan)
		close(portChan)
		close(done)
	}()

	go flow.FetchExternalIP(nc, &wg, ipChan, done)
	go flow.MapPorts(nc, 600, &wg, ipChan, portChan, done)

	select {
	case p := <-portChan:
		t.Fatalf("Should not have been able to open a port, opened %d instead?", p)
	case <-time.After(5 * time.Second):
		done <- os.Interrupt
	}

	wg.Wait()

	assert.True(t, externalIPRetrieved)
	assert.False(t, peerPortSet)
	assert.False(t, portMapped)
}

// NAT is OK, but Transmission is not connected
func Test_Flow_NoTransmissionConnection(t *testing.T) {
	var wg sync.WaitGroup
	reset()
	wg.Add(3)

	nc := &fakeNatClient{true, true}
	tc := &fakeTransmissionClient{false, false, false}

	ipChan := make(chan string)
	portChan := make(chan int)
	done := make(chan os.Signal, 1)
	defer func() {
		close(ipChan)
		close(portChan)
		close(done)
	}()

	// Start the flow
	go flow.FetchExternalIP(nc, &wg, ipChan, done)
	go flow.MapPorts(nc, 600, &wg, ipChan, portChan, done)
	go flow.TransmissionArgSetter(tc, &wg, portChan, done)

	// Wait for everyone to finish
	<-time.After(5 * time.Second)
	done <- os.Interrupt
	wg.Wait()

	assert.True(t, externalIPRetrieved)
	assert.True(t, portMapped)
	assert.False(t, peerPortSet)
}

// NAT is OK, Transmission is connected but the port is already open
func Test_Flow_TransmissionPortAlreadyOpen(t *testing.T) {
	var wg sync.WaitGroup
	reset()
	wg.Add(3)

	nc := &fakeNatClient{true, true}
	tc := &fakeTransmissionClient{true, true, true}

	ipChan := make(chan string)
	portChan := make(chan int)
	done := make(chan os.Signal, 1)
	defer func() {
		close(ipChan)
		close(portChan)
		close(done)
	}()

	// Start the flow
	go flow.FetchExternalIP(nc, &wg, ipChan, done)
	go flow.MapPorts(nc, 600, &wg, ipChan, portChan, done)
	go flow.TransmissionArgSetter(tc, &wg, portChan, done)

	// Wait for everyone to finish
	<-time.After(5 * time.Second)
	done <- os.Interrupt
	wg.Wait()

	assert.True(t, externalIPRetrieved)
	assert.True(t, portMapped)
	assert.False(t, peerPortSet)
}

// NAT is OK, Transmission is connected, the port is not open but we cannot set it
func Test_Flow_TransmissionUnableToSet(t *testing.T) {
	var wg sync.WaitGroup
	reset()
	wg.Add(3)

	nc := &fakeNatClient{true, true}
	tc := &fakeTransmissionClient{true, false, false}

	ipChan := make(chan string)
	portChan := make(chan int)
	done := make(chan os.Signal, 1)
	defer func() {
		close(ipChan)
		close(portChan)
		close(done)
	}()

	// Start the flow
	go flow.FetchExternalIP(nc, &wg, ipChan, done)
	go flow.MapPorts(nc, 600, &wg, ipChan, portChan, done)
	go flow.TransmissionArgSetter(tc, &wg, portChan, done)

	// Wait for everyone to finish
	<-time.After(5 * time.Second)
	done <- os.Interrupt
	wg.Wait()

	assert.True(t, externalIPRetrieved)
	assert.True(t, portMapped)
	assert.False(t, peerPortSet)
}
