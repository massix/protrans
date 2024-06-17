package transmission

import (
	"github.com/hekmon/transmissionrpc"
)

type Client interface {
	IsConnected() bool
	IsPortOpen() bool
	GetCurrentPort() (int, error)
	SetPeerPort(port int) error
}

type TransmissionRpcClient interface {
	PortTest() (bool, error)
	SessionStats() (*transmissionrpc.SessionStats, error)
	SessionArgumentsSet(*transmissionrpc.SessionArguments) error
	SessionArgumentsGet() (*transmissionrpc.SessionArguments, error)
}

type encapsulatedClient struct {
	TransmissionRpcClient
}

func New(client TransmissionRpcClient) Client {
	return &encapsulatedClient{client}
}

// GetCurrentPort implements Client.
func (ec *encapsulatedClient) GetCurrentPort() (int, error) {
	res, err := ec.SessionArgumentsGet()
	if err != nil {
		return 0, err
	}

	return int(*res.PeerPort), nil
}

// IsConnected implements Client.
func (ec *encapsulatedClient) IsConnected() bool {
	_, err := ec.SessionStats()
	return err == nil
}

// IsPortOpen implements Client.
func (ec *encapsulatedClient) IsPortOpen() bool {
	res, err := ec.PortTest()
	return res && err == nil
}

// SetPeerPort implements Client.
func (ec *encapsulatedClient) SetPeerPort(port int) error {
	newPort := int64(port)
	return ec.SessionArgumentsSet(&transmissionrpc.SessionArguments{
		PeerPort: &newPort,
	})
}
