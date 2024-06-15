package config_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/massix/protrans/internal/config"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func Test_ShouldParseConfiguration(t *testing.T) {
	os.Setenv("PROTRANS_TRANSMISSION_PORT", "1234")
	os.Setenv("PROTRANS_NAT_GATEWAY", "192.168.1.1")

	res := config.New("../../tests/test_configuration.yaml", true)
	assert.NotNil(t, res)
	assert.Equal(t, "localhost", res.Transmission.Host)              // Default value
	assert.Equal(t, uint16(1234), res.Transmission.Port)             // Overridden by environment
	assert.Equal(t, "from_configuration", res.Transmission.Username) // From configuration file
	assert.Equal(t, "192.168.1.1", res.Nat.Gateway)                  // Overridden by environment
	assert.Equal(t, uint16(600), res.Nat.PortLifetime)               // Default value
	assert.Equal(t, logrus.WarnLevel, res.LogrusLogLevel())          // From configuration file
}

func Test_ShouldOnlyUseEnvironment(t *testing.T) {
	os.Setenv("PROTRANS_NAT_GATEWAY", "192.168.42.1")
	res := config.New("", true)
	assert.Equal(t, "localhost", res.Transmission.Host)
	assert.Equal(t, "192.168.42.1", res.Nat.Gateway)
}

func Test_FileDoesNotExist_DefaultConfiguration(t *testing.T) {
	res := config.New("not-exists.yaml", false)
	assert.Equal(t, "localhost", res.Transmission.Host)
}

func Test_InvalidYAMLFile_DefaultConfiguration(t *testing.T) {
	res := config.New("../../tests/test_invalid_configuration.yaml", false)
	assert.Equal(t, uint16(9091), res.Transmission.Port)
}

func Test_InvalidFile_UseEnvironment(t *testing.T) {
	os.Setenv("PROTRANS_TRANSMISSION_USERNAME", "user_from_environment")
	res := config.New("../../tests/test_invalid_configuration.yaml", true)
	assert.Equal(t, "user_from_environment", res.Transmission.Username)
}

func Test_GatewayIP_OK(t *testing.T) {
	conf := &config.ProtransConfiguration{Nat: &config.NatConfiguration{Gateway: "192.168.42.1"}}
	res := conf.GatewayIP()

	assert.Equal(t, "192.168.42.1", res.String())
}

func Test_GatewayIP_KO(t *testing.T) {
	conf := &config.ProtransConfiguration{Nat: &config.NatConfiguration{Gateway: "Invalid gateway"}}
	res := conf.GatewayIP()

	assert.Equal(t, "10.2.0.1", res.String())
}

func ExampleTransmissionConfiguration_String() {
	tc := &config.TransmissionConfiguration{
		Host:     "localhost",
		Port:     9091,
		Username: "SomeUser",
		Password: "WithPassword",
	}

	fmt.Print(tc.String())

	// Output:
	// Transmission{Host="localhost" Port=9091 Username="SomeUser" Password="set"}
}

func ExampleNatConfiguration_String() {
	nc := &config.NatConfiguration{
		Gateway:      "10.0.2.1",
		PortLifetime: 600,
	}

	fmt.Print(nc)

	// Output:
	// Nat{Gateway="10.0.2.1" PortLifetime=600}
}

func ExampleProtransConfiguration_String() {
	ptc := &config.ProtransConfiguration{
		Transmission: &config.TransmissionConfiguration{
			Host:     "192.168.1.2",
			Port:     8080,
			Username: "",
			Password: "",
		},
		Nat: &config.NatConfiguration{
			Gateway:      "10.42.3.1",
			PortLifetime: 126,
		},
		LogLevel: "DEBUG",
	}

	fmt.Print(ptc)

	// Output:
	// ProTrans{LogLevel="DEBUG" Nat{Gateway="10.42.3.1" PortLifetime=126} Transmission{Host="192.168.1.2" Port=8080 Username="" Password="not set"}}
}
