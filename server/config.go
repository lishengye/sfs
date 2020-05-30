package server

import (
	"encoding/json"
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
	return config, nil
}
