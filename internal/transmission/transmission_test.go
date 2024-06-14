package transmission_test

import (
	"errors"
	"testing"

	"github.com/hekmon/transmissionrpc"
	"github.com/massix/protrans/internal/transmission"
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

func newFakeClient(isConnected, isPortOpen bool) transmission.Client {
	return transmission.New(&MockClient{isConnected, isPortOpen})
}

func Test_IsConnected_OK(t *testing.T) {
	cl := newFakeClient(true, true)
	res := cl.IsConnected()
	assert.True(t, res)
}

func Test_IsConnected_Fail(t *testing.T) {
	cl := newFakeClient(false, false)
	res := cl.IsConnected()
	assert.False(t, res)
}

func Test_IsPortOpen_OK(t *testing.T) {
	cl := newFakeClient(true, true)
	res := cl.IsPortOpen()
	assert.True(t, res)
}

func Test_IsPortOpen_FailNotConnected(t *testing.T) {
	cl := newFakeClient(false, true)
	res := cl.IsPortOpen()
	assert.False(t, res)
}

func Test_IsPortOpen_FailNotOpen(t *testing.T) {
	cl := newFakeClient(true, false)
	res := cl.IsPortOpen()
	assert.False(t, res)
}

func Test_GetCurrentPort_OK(t *testing.T) {
	cl := newFakeClient(true, false)
	res, err := cl.GetCurrentPort()
	assert.Nil(t, err)
	assert.Equal(t, 4242, res)
}

func Test_GetCurrentPort_FailNotConnected(t *testing.T) {
	cl := newFakeClient(false, false)
	res, err := cl.GetCurrentPort()
	assert.Equal(t, 0, res)
	assert.Error(t, err)
	assert.Equal(t, err.Error(), "Client not connected")
}

func Test_SetPeerPort_OK(t *testing.T) {
	cl := newFakeClient(true, true)
	err := cl.SetPeerPort(4242)
	assert.Nil(t, err)
}

func Test_SetPeerPort_FailNotConnected(t *testing.T) {
	cl := newFakeClient(false, false)
	err := cl.SetPeerPort(4242)
	assert.Error(t, err)
	assert.Equal(t, err.Error(), "Client not connected")
}

func Test_SetPeerPort_FailInvalidPort(t *testing.T) {
	cl := newFakeClient(true, false)
	err := cl.SetPeerPort(42)
	assert.Error(t, err)
	assert.Equal(t, err.Error(), "Invalid port number")
}
