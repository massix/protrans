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

type EncapsulatedClient struct {
	client TransmissionRpcClient
}

func New(client TransmissionRpcClient) Client {
	return &EncapsulatedClient{client}
}

// GetCurrentPort implements Client.
func (e *EncapsulatedClient) GetCurrentPort() (int, error) {
	res, err := e.client.SessionArgumentsGet()
	if err != nil {
		return 0, err
	}

	return int(*res.PeerPort), nil
}

// IsConnected implements Client.
func (e *EncapsulatedClient) IsConnected() bool {
	_, err := e.client.SessionStats()
	return err == nil
}

// IsPortOpen implements Client.
func (e *EncapsulatedClient) IsPortOpen() bool {
	res, err := e.client.PortTest()
	return res && err == nil
}

// SetPeerPort implements Client.
func (e *EncapsulatedClient) SetPeerPort(port int) error {
	newPort := int64(port)
	return e.client.SessionArgumentsSet(&transmissionrpc.SessionArguments{
		PeerPort: &newPort,
	})
}
