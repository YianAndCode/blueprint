package main

import (
	"encoding/json"
	"os"
	"path"
)

type DBConfig struct {
	Type string `json:"type"`
	Host string `json:"host"`
	Port uint   `json:"port"`
	User string `json:"user"`
	Pass string `json:"pass"`
	Name string `json:"name"`
}

type Config struct {
	Env       string     `json:"env"`
	Databases []DBConfig `json:"databases"`
}

var config Config

func loadJsonConfig(filepath string) error {
	configFile, err := os.ReadFile(path.Join(filepath, BlueprintConfigFileName))
	if err != nil {
		return err
	}

	err = json.Unmarshal(configFile, &config)
	if err != nil {
		return err
	}

	// Set default type to mysql for backward compatibility
	for i := range config.Databases {
		if config.Databases[i].Type == "" {
			config.Databases[i].Type = "mysql"
		}
	}

	return nil
}
