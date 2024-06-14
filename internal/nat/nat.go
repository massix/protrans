package nat

import (
	"fmt"

	natpmp "github.com/jackpal/go-nat-pmp"
)

func ipToString(ip [4]byte) string {
	return fmt.Sprintf("%d.%d.%d.%d", ip[0], ip[1], ip[2], ip[3])
}

// This will make it easier to mock the tests
type Client interface {
	GetExternalAddress() (string, error)
	AddPortMapping(protocol string, lifetime int) (int, error)
}

type NatpmpClient interface {
	AddPortMapping(protocol string, internalPort, externalPort, lifetime int) (*natpmp.AddPortMappingResult, error)
	GetExternalAddress() (*natpmp.GetExternalAddressResult, error)
}

type EncapsulatedClient struct {
	Client NatpmpClient
}

func New(client NatpmpClient) Client {
	return &EncapsulatedClient{client}
}

// AddPortMapping implements Client.
func (e *EncapsulatedClient) AddPortMapping(protocol string, lifetime int) (int, error) {
	res, err := e.Client.AddPortMapping(protocol, 0, 1, lifetime)
	if err != nil {
		return 0, &PortMappingUnavailableError{}
	}

	return int(res.MappedExternalPort), nil
}

// GetExternalAddress implements Client.
func (e *EncapsulatedClient) GetExternalAddress() (string, error) {
	res, err := e.Client.GetExternalAddress()
	if err != nil {
		return "", &ErrClientNotConnected{}
	}

	return ipToString(res.ExternalIPAddress), nil
}

type ErrClientNotConnected struct{}

func (n *ErrClientNotConnected) Error() string {
	return "NAT Client is not connected"
}

type PortMappingUnavailableError struct{}

func (n *PortMappingUnavailableError) Error() string {
	return "Failure when trying to map port"
}
