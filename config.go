package main

import (
	"encoding/json"
	"os"
	"path"
)

type DBType string

const (
	MySQL  DBType = "mysql"
	PG     DBType = "pg"
	SQLite DBType = "sqlite"
)

func (t DBType) IsValid() bool {
	switch t {
	case MySQL, PG, SQLite:
		return true
	}
	return false
}

type DBConfig struct {
	Type DBType `json:"type"` // used by all
	Host string `json:"host"` // used by mysql, pg
	Port uint   `json:"port"` // used by mysql, pg
	User string `json:"user"` // used by mysql, pg
	Pass string `json:"pass"` // used by mysql, pg
	Name string `json:"name"` // used by mysql, pg
	File string `json:"file"` // used by sqlite
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
			config.Databases[i].Type = MySQL
		}
	}

	return nil
}
