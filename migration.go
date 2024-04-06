package main

import (
	"os"
)

type MigrationInfo struct {
	Id        uint
	Migration string
	Batch     uint
}

type Migration struct {
	Name         string
	UpFilename   string
	DownFilename string

	upSQL   string
	downSQL string
}

func (m *Migration) LoadSQLFile() error {
	up, err := os.ReadFile(m.UpFilename)
	if err != nil {
		return err
	}
	m.upSQL = string(up)

	down, err := os.ReadFile(m.DownFilename)
	if err != nil {
		return err
	}
	m.downSQL = string(down)

	return nil
}

func (m Migration) GetUpSQL() string {
	return m.upSQL
}

func (m Migration) GetDownSQL() string {
	return m.downSQL
}
