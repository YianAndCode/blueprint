package main

import (
	"encoding/json"
	"os"
	"path"
)

type MySQLConfig struct {
	Host string `json:"host"`
	Port uint   `json:"port"`
	User string `json:"user"`
	Pass string `json:"pass"`
	Name string `json:"name"`
}

type Config struct {
	Env       string        `json:"env"`
	Databases []MySQLConfig `json:"databases"`
}

var config Config

func loadJsonConfig(filepath string) error {
	configFile, err := os.ReadFile(path.Join(filepath, BlueprintConfigFileName))
	if err != nil {
		return err
	}

	return json.Unmarshal(configFile, &config)
}
