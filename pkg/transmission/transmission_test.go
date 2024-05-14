package transmission_test

import (
	"errors"
	"testing"

	"github.com/hekmon/transmissionrpc"
	"github.com/massix/protrans/pkg/transmission"
	"github.com/stretchr/testify/assert"
)

type MockClient struct {
	IsConnected bool
	IsPortOpen  bool
}

type ClientNotConnectedError struct{}

func (c *ClientNotConnectedError) Error() string {
	return "Client not connected"
}

func (m *MockClient) SessionArgumentsGet() (*transmissionrpc.SessionArguments, error) {
	port := int64(4242)

	if !m.IsConnected {
		return nil, &ClientNotConnectedError{}
	}

	return &transmissionrpc.SessionArguments{PeerPort: &port}, nil
}

func (m *MockClient) SessionArgumentsSet(args *transmissionrpc.SessionArguments) error {
	if !m.IsConnected {
		return &ClientNotConnectedError{}
	}

	if *args.PeerPort < 1024 {
		return errors.New("Invalid port number")
	}

	return nil
}

func (m *MockClient) PortTest() (bool, error) {
	if !m.IsConnected {
		return false, &ClientNotConnectedError{}
	}

	return m.IsPortOpen, nil
}

func (m *MockClient) SessionStats() (*transmissionrpc.SessionStats, error) {
	if !m.IsConnected {
		return nil, &ClientNotConnectedError{}
	}

	return &transmissionrpc.SessionStats{}, nil
}

func Test_IsConnected_OK(t *testing.T) {
	res := transmission.IsConnected(&MockClient{true, true})
	assert.True(t, res)
}

func Test_IsConnected_Fail(t *testing.T) {
	res := transmission.IsConnected(&MockClient{false, false})
	assert.False(t, res)
}

func Test_IsPortOpen_OK(t *testing.T) {
	res := transmission.IsPortOpen(&MockClient{true, true})
	assert.True(t, res)
}

func Test_IsPortOpen_FailNotConnected(t *testing.T) {
	res := transmission.IsPortOpen(&MockClient{false, true})
	assert.False(t, res)
}

func Test_IsPortOpen_FailNotOpen(t *testing.T) {
	res := transmission.IsPortOpen(&MockClient{true, false})
	assert.False(t, res)
}

func Test_GetCurrentPort_OK(t *testing.T) {
	res, err := transmission.GetCurrentPort(&MockClient{true, false})
	assert.Nil(t, err)
	assert.Equal(t, 4242, res)
}

func Test_GetCurrentPort_FailNotConnected(t *testing.T) {
	res, err := transmission.GetCurrentPort(&MockClient{false, false})
	assert.Equal(t, 0, res)
	assert.Error(t, err)
	assert.Equal(t, err.Error(), "Client not connected")
}

func Test_SetPeerPort_OK(t *testing.T) {
	err := transmission.SetPeerPort(&MockClient{true, true}, 4242)
	assert.Nil(t, err)
}

func Test_SetPeerPort_FailNotConnected(t *testing.T) {
	err := transmission.SetPeerPort(&MockClient{false, false}, 4242)
	assert.Error(t, err)
	assert.Equal(t, err.Error(), "Client not connected")
}

func Test_SetPeerPort_FailInvalidPort(t *testing.T) {
	err := transmission.SetPeerPort(&MockClient{true, false}, 42)
	assert.Error(t, err)
	assert.Equal(t, err.Error(), "Invalid port number")
}
