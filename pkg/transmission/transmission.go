package transmission

import (
	"github.com/hekmon/transmissionrpc"
)

type TransmissionClient interface {
	PortTest() (bool, error)
	SessionStats() (*transmissionrpc.SessionStats, error)
	SessionArgumentsSet(*transmissionrpc.SessionArguments) error
	SessionArgumentsGet() (*transmissionrpc.SessionArguments, error)
}

func IsConnected(c TransmissionClient) bool {
	_, err := c.SessionStats()
	return err == nil
}

func IsPortOpen(c TransmissionClient) bool {
	res, err := c.PortTest()
	return res && err == nil
}

func GetCurrentPort(c TransmissionClient) (int, error) {
	res, err := c.SessionArgumentsGet()
	if err != nil {
		return 0, err
	}

	return int(*res.PeerPort), nil
}

func SetPeerPort(c TransmissionClient, port int) error {
	newPort := int64(port)
	return c.SessionArgumentsSet(&transmissionrpc.SessionArguments{
		PeerPort: &newPort,
	})
}
