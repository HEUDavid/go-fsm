package util

import (
	"github.com/BurntSushi/toml"
	"path/filepath"
)

type Config = map[string]any

func loadConfig(file string) (*Config, error) {
	var conf Config
	if _, err := toml.DecodeFile(file, &conf); err != nil {
		return nil, err
	}
	return &conf, nil
}

var config *Config

func GetConfig() *Config {
	if config == nil {
		conf, err := loadConfig(filepath.Join(FindProjectRoot(), "conf", "conf.toml"))
		if err != nil {
			panic(err)
		}
		config = conf
	}
	return config
}
