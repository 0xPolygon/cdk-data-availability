package config

import (
	"bytes"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

// DefaultValues is the default configuration
const DefaultValues = `
PrivateKey = {Path = "/pk/test-member.keystore", Password = "testonly"}

[L1]
WsURL = "ws://127.0.0.1:8546"
RpcURL = "http://127.0.0.1:8545"
PolygonValidiumAddress = "0x8dAF17A20c9DBA35f005b6324F493785D239719d"
DataCommitteeAddress = "0x68B1D87F95878fE05B998F19b66F4baba5De1aed"
Timeout = "1m"
RetryPeriod = "5s"
BlockBatchSize = "64"

[Log]
Environment = "development" # "production" or "development"
Level = "info"
Outputs = ["stderr"]

[DB]
User = "committee_user"
Password = "committee_password"
Name = "committee_db"
Host = "cdk-data-availability-db"
Port = "5432"
EnableLog = false
MaxConns = 200

[RPC]
Host = "0.0.0.0"
Port = 8444
ReadTimeout = "60s"
WriteTimeout = "60s"
MaxRequestsPerIPAndSecond = 500
`

// Default parses the default configuration values.
func Default() (*Config, error) {
	var cfg Config
	viper.SetConfigType("toml")

	err := viper.ReadConfig(bytes.NewBuffer([]byte(DefaultValues)))
	if err != nil {
		return nil, err
	}
	err = viper.Unmarshal(&cfg, viper.DecodeHook(mapstructure.TextUnmarshallerHookFunc()))
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
