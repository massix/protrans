package nat

import (
	"fmt"

	natpmp "github.com/jackpal/go-nat-pmp"
)

// This will make it easier to mock the tests
type NatClientI interface {
	GetExternalAddress() (*natpmp.GetExternalAddressResult, error)
	AddPortMapping(protocol string, internalPort, requestedExternalPort int, lifetime int) (*natpmp.AddPortMappingResult, error)
}

type NatClientNotConnectedError struct{}

func (n *NatClientNotConnectedError) Error() string {
	return "NAT Client is not connected"
}

type PortMappingUnavailableError struct{}

func (n *PortMappingUnavailableError) Error() string {
	return "Failure when trying to map port"
}

func ipToString(ip [4]byte) string {
	return fmt.Sprintf("%d.%d.%d.%d", ip[0], ip[1], ip[2], ip[3])
}

func GetExternalIP(client NatClientI) (string, error) {
	res, err := client.GetExternalAddress()
	if err != nil {
		return "", &NatClientNotConnectedError{}
	}

	return ipToString(res.ExternalIPAddress), nil
}

func AddPortMapping(client NatClientI, protocol string, portLifeTime int) (int, error) {
	res, err := client.AddPortMapping(protocol, 0, 1, portLifeTime)
	if err != nil {
		return 0, &PortMappingUnavailableError{}
	}

	return int(res.MappedExternalPort), nil
}
