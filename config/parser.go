package config

import (
	"errors"
	"net/url"
	"strconv"
)

var missingParams = "You need to provide <local port> <proxy host:port> <true|false> (https)"
var missingHttpsConfig = missingParams + " <cert path> <key path>"
var invalidPort = "Invalid port number"
var invalidBool = "Invalid boolean value for https"

type Config struct {
	LocalPort   int
	ProxyUrl    *url.URL
	HttpsConfig HttpsConfig
}

type HttpsConfig struct {
	Enabled  bool
	CertPath string
	KeyPath  string
}

func ParseConfig(args []string) (Config, error) {
	if len(args) < 3 {
		return Config{}, errors.New(missingParams)
	}

	port64, err := strconv.ParseInt(args[0], 10, 0)
	if err != nil {
		return Config{}, errors.New(invalidPort)
	}

	port32 := int32(port64)

	https, err := strconv.ParseBool(args[2])
	if err != nil {
		return Config{}, errors.New(invalidBool)
	}

	proxyUrl, err := url.Parse(args[1])
	if err != nil {
		return Config{}, err
	}

	httpsConfig, err := parseHttpsConfig(https, args)
	if err != nil {
		return Config{}, err
	}

	return Config{
		LocalPort:   int(port32),
		ProxyUrl:    proxyUrl,
		HttpsConfig: httpsConfig,
	}, nil
}

func parseHttpsConfig(isHttps bool, args []string) (HttpsConfig, error) {
	if isHttps {
		if len(args) < 5 {
			return HttpsConfig{}, errors.New(missingHttpsConfig)
		}

		return HttpsConfig{
			Enabled:  true,
			CertPath: args[3],
			KeyPath:  args[4],
		}, nil
	} else {
		return HttpsConfig{
			Enabled: false,
		}, nil
	}
}
