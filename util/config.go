package util

import (
	"time"

	"github.com/spf13/viper"
)

/*
* Config stores the all configuration off the application
* The configurations are read by viper from a config file or env file
 */
type Config struct {
	DBDriver            string        `mapstructure:"DB_DRIVER"`
	DBSource            string        `mapstructure:"DB_SOURCE"`
	ServerAddress       string        `mapstructure:"SERVER_ADDRESS"`
	TokenSymmetricKey   string        `mapstructure:"TOKEN_SYMMETRIC_KEY"`
	AccessTokenDuration time.Duration `mapstructure:"ACCESS_TOKEN_DURATION"`
	TokenType           string        `mapstructure:"TOKEN_TYPE"`
}

// LoadConfig - reads the conf file ot the env file
func LoadConfig(path string) (config Config, err error) {
	// setting the config file folder path
	viper.AddConfigPath(path)
	// setting the name of the config file
	viper.SetConfigName(".app")
	// setting the config type as enf file
	viper.SetConfigType("env")
	// overwrite environment variables from the file with corresponding values if they already exist.
	viper.AutomaticEnv()
	// read the configurations from the disk
	if err = viper.ReadInConfig(); err != nil {
		return
	}
	// extracting the configurations into the config type
	err = viper.Unmarshal(&config)
	return
}
