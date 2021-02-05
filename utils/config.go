package utils

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
}

func init() {
	viper.SetConfigName("task")
	viper.AddConfigPath("../timer-task-manager/config/")
	//viper.AddConfigPath("$HOME/config/")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal(err)
	}
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) Get(key string) interface{} {
	return viper.Get(key)
}
