package server

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	Port      int16
	Directory string
}

func NewConfig(configFile string) (Config, error) {
	file, err := os.Open(configFile)
	config := Config{}
	if err != nil {
		return config, err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config);err != nil {
		return config, err
	}

	if err := config.check(); err != nil {
		return config, err
	}

	return config, nil
}

func (config *Config) check() error {
	file, err := os.Stat(config.Directory)
	if err != nil {
		return fmt.Errorf("Config error: %s not exist", config.Directory)
	}
	if !file.IsDir() {
		return fmt.Errorf("Config error: %s not directory", file.Name())
	}
	return nil
}
