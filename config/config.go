package config

import (
	"crypto/ecdsa"
	"os"
	"path/filepath"
	"strings"

	"github.com/0xPolygon/cdk-data-availability/config/types"
	"github.com/0xPolygon/cdk-data-availability/db"
	"github.com/0xPolygon/cdk-data-availability/log"
	"github.com/0xPolygon/cdk-data-availability/rpc"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"github.com/urfave/cli/v2"
)

const (
	// FlagCfg flag used for config aka cfg
	FlagCfg = "cfg"
)

// Config represents the full configuration of the data node
type Config struct {
	PrivateKey types.KeystoreFileConfig
	DB         db.Config
	Log        log.Config
	RPC        rpc.Config
	L1         L1Config
}

// L1Config is a struct that defines L1 contract and service settings
type L1Config struct {
	WsURL                  string         `mapstructure:"WsURL"`
	RpcURL                 string         `mapstructure:"RpcURL"`
	PolygonValidiumAddress string         `mapstructure:"PolygonValidiumAddress"`
	DataCommitteeAddress   string         `mapstructure:"DataCommitteeAddress"`
	Timeout                types.Duration `mapstructure:"Timeout"`
	RetryPeriod            types.Duration `mapstructure:"RetryPeriod"`
	BlockBatchSize         uint           `mapstructure:"BlockBatchSize"`
}

// Load loads the configuration baseed on the cli context
func Load(ctx *cli.Context) (*Config, error) {
	cfg, err := Default()
	if err != nil {
		return nil, err
	}

	viper.AutomaticEnv()
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)
	viper.SetEnvPrefix("DATA_NODE")

	configFilePath := ctx.String(FlagCfg)
	if configFilePath != "" {
		dirName, fileName := filepath.Split(configFilePath)

		fileExtension := strings.TrimPrefix(filepath.Ext(fileName), ".")
		fileNameWithoutExtension := strings.TrimSuffix(fileName, "."+fileExtension)

		viper.AddConfigPath(dirName)
		viper.SetConfigName(fileNameWithoutExtension)
		viper.SetConfigType(fileExtension)
		err = viper.ReadInConfig()
		if err != nil {
			return nil, err
		}
	}

	decodeHooks := []viper.DecoderConfigOption{
		// this allows arrays to be decoded from env var separated by ",", example: MY_VAR="value1,value2,value3"
		viper.DecodeHook(mapstructure.ComposeDecodeHookFunc(mapstructure.TextUnmarshallerHookFunc(), mapstructure.StringToSliceHookFunc(","))),
	}
	err = viper.Unmarshal(&cfg, decodeHooks...)
	return cfg, err
}

// NewKeyFromKeystore creates a private key from a keystore file
func NewKeyFromKeystore(cfg types.KeystoreFileConfig) (*ecdsa.PrivateKey, error) {
	if cfg.Path == "" && cfg.Password == "" {
		return nil, nil
	}
	keystoreEncrypted, err := os.ReadFile(filepath.Clean(cfg.Path))
	if err != nil {
		return nil, err
	}
	key, err := keystore.DecryptKey(keystoreEncrypted, cfg.Password)
	if err != nil {
		return nil, err
	}
	return key.PrivateKey, nil
}
