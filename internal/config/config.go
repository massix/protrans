package config

import (
	"fmt"
	"net"
	"os"
	"reflect"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type (
	ProtransConfiguration struct {
		Transmission *TransmissionConfiguration `yaml:"transmission"`
		Nat          *NatConfiguration          `yaml:"nat"`
		LogLevel     string                     `yaml:"log_level" env:"PROTRANS_LOG_LEVEL"`
	}

	TransmissionConfiguration struct {
		Host     string `yaml:"host" env:"PROTRANS_TRANSMISSION_HOST"`
		Port     uint16 `yaml:"port" env:"PROTRANS_TRANSMISSION_PORT"`
		Username string `yaml:"username" env:"PROTRANS_TRANSMISSION_USERNAME"`
		Password string `yaml:"password" env:"PROTRANS_TRANSMISSION_PASSWORD"`
	}

	NatConfiguration struct {
		Gateway      string `yaml:"gateway" env:"PROTRANS_NAT_GATEWAY"`
		PortLifetime uint16 `yaml:"port_lifetime" env:"PROTRANS_NAT_PORT_LIFETIME"`
	}
)

func overrideWithEnvironment(iface, concrete any) {
	types := reflect.TypeOf(iface)
	values := reflect.ValueOf(concrete).Elem()

	for i := range types.NumField() {
		typeField := types.Field(i)
		valueField := values.FieldByName(typeField.Name)
		if env, ok := typeField.Tag.Lookup("env"); ok && valueField.CanSet() && valueField.IsValid() {
			logrus.Debugf("found env: %s", env)
			if stringValue, ok := os.LookupEnv(env); ok {
				switch valueField.Type().Kind() {
				case reflect.String:
					logrus.Debugf("Environment variable: %s=%s", env, stringValue)
					valueField.SetString(stringValue)
				case reflect.Uint16:
					var newValue uint16
					fmt.Sscanf(stringValue, "%d", &newValue)
					logrus.Debugf("Environment variable: %s=%d", env, newValue)
					valueField.SetUint(uint64(newValue))
				}
			}
		}
	}
}

func New(filePath string, shouldUseEnvironment bool) *ProtransConfiguration {
	// Init with some default values
	conf := &ProtransConfiguration{
		Transmission: &TransmissionConfiguration{
			Host: "localhost",
			Port: 9091,
		},
		Nat: &NatConfiguration{
			Gateway:      "10.2.0.1",
			PortLifetime: 120,
		},

		LogLevel: "INFO",
	}

	if filePath != "" {
		content, err := os.ReadFile(filePath)
		if err != nil {
			logrus.Errorf("Failure when reading file: '%s' %s", filePath, err)
		}

		err = yaml.Unmarshal(content, &conf)
		if err != nil {
			logrus.Errorf("Failure when unmarshalling file '%s': %s", filePath, err)
		}
	}

	// Environment Variables should override the values from the YAML file
	if shouldUseEnvironment {
		if logLevel, ok := os.LookupEnv("PROTRANS_LOG_LEVEL"); ok {
			conf.LogLevel = logLevel
		}

		logrus.Infof("Using environment variables")

		overrideWithEnvironment(TransmissionConfiguration{}, conf.Transmission)
		overrideWithEnvironment(NatConfiguration{}, conf.Nat)
	}

	return conf
}

func (nc *NatConfiguration) String() string {
	return fmt.Sprintf("Nat{Gateway=%q PortLifetime=%d}", nc.Gateway, nc.PortLifetime)
}

func (tc *TransmissionConfiguration) String() string {
	password := "not set"
	if tc.Password != "" {
		password = "set"
	}

	return fmt.Sprintf("Transmission{Host=%q Port=%d Username=%q Password=%q}", tc.Host, tc.Port, tc.Username, password)
}

func (pc *ProtransConfiguration) String() string {
	return fmt.Sprintf("ProTrans{LogLevel=%q %s %s}", pc.LogLevel, pc.Nat, pc.Transmission)
}

func (c *ProtransConfiguration) LogrusLogLevel() (level logrus.Level) {
	level = logrus.InfoLevel
	switch c.LogLevel {
	case "TRACE":
		level = logrus.TraceLevel
	case "DEBUG":
		level = logrus.DebugLevel
	case "INFO":
		level = logrus.InfoLevel
	case "WARN":
		level = logrus.WarnLevel
	case "ERROR":
		level = logrus.ErrorLevel
	case "FATAL":
		level = logrus.FatalLevel
	case "PANIC":
		level = logrus.PanicLevel
	}

	return
}

func (c *ProtransConfiguration) GatewayIP() net.IP {
	var bytes [4]byte
	_, err := fmt.Sscanf(c.Nat.Gateway, "%d.%d.%d.%d", &bytes[0], &bytes[1], &bytes[2], &bytes[3])
	if err != nil {
		logrus.Warnf("Could not parse NAT Gateway IP '%s': %s, using default value", c.Nat.Gateway, err)
		return net.IPv4(10, 2, 0, 1)
	}

	return net.IPv4(bytes[0], bytes[1], bytes[2], bytes[3])
}
