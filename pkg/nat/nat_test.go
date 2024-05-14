package nat_test

import (
	"errors"
	"testing"

	natpmp "github.com/jackpal/go-nat-pmp"
	"github.com/massix/protrans/pkg/nat"
	"github.com/stretchr/testify/assert"
)

type MockClient struct {
	ExternalAddressShouldFail bool
	PortMappingShouldFail     bool
}

func (m *MockClient) GetExternalAddress() (*natpmp.GetExternalAddressResult, error) {
	if m.ExternalAddressShouldFail {
		return nil, errors.New("Failure")
	}

	return &natpmp.GetExternalAddressResult{
		SecondsSinceStartOfEpoc: 0,
		ExternalIPAddress:       [4]byte{10, 20, 30, 40},
	}, nil
}

func (m *MockClient) AddPortMapping(protocol string, par1, par2, par3 int) (*natpmp.AddPortMappingResult, error) {
	if protocol != "tcp" && protocol != "udp" {
		return nil, errors.New("Wrong protocol")
	}

	if m.PortMappingShouldFail {
		return nil, errors.New("Failure")
	}

	return &natpmp.AddPortMappingResult{
		SecondsSinceStartOfEpoc:      0,
		MappedExternalPort:           1234,
		InternalPort:                 1234,
		PortMappingLifetimeInSeconds: 600,
	}, nil
}

func Test_GetExternalIP_OK(t *testing.T) {
	res, err := nat.GetExternalIP(&MockClient{false, false})
	assert.Nil(t, err)
	assert.Equal(t, "10.20.30.40", res)
}

func Test_GetExternalIP_Fail(t *testing.T) {
	res, err := nat.GetExternalIP(&MockClient{true, false})
	assert.Emptyf(t, res, "IP Address should be empty")
	assert.Errorf(t, err, "NAT Client is not connected")
}

func Test_PortMapping_OK(t *testing.T) {
	res, err := nat.AddPortMapping(&MockClient{false, false}, "tcp", 30)
	assert.Nil(t, err)
	assert.Equal(t, 1234, res)
}

func Test_PortMapping_FailProtocol(t *testing.T) {
	res, err := nat.AddPortMapping(&MockClient{false, false}, "unknown", 30)
	assert.Empty(t, res)
	assert.Errorf(t, err, "Wrong protocol")
}

func Test_PortMapping_Fail(t *testing.T) {
	res, err := nat.AddPortMapping(&MockClient{false, true}, "udp", 30)
	assert.Empty(t, res)
	assert.Errorf(t, err, "Failure when trying to map port")
}
