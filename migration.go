package main

import (
	"os"
	"path"
	"strings"
)

// 数据库 Migration 表结构
type MigrationRec struct {
	Id        uint
	Migration string
	Batch     uint
}

type MigrationInfo struct {
	Name         string
	UpFilename   string
	DownFilename string

	upSQL   string
	downSQL string
}

func (m *MigrationInfo) LoadSQLFile() error {
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

func (m MigrationInfo) GetUpSQL() string {
	return m.upSQL
}

func (m MigrationInfo) GetDownSQL() string {
	return m.downSQL
}

type Migrations struct {
	names []string
	infos map[string]MigrationInfo
}

func (m *Migrations) GetNames() []string {
	return m.names
}

func (m *Migrations) GetInfo(name string) MigrationInfo {
	return m.infos[name]
}

func LoadMigrations(migrationPath string) (*Migrations, error) {
	dirs, err := os.ReadDir(migrationPath)
	if err != nil {
		return nil, err
	}

	migrations := &Migrations{
		names: make([]string, 0),
		infos: make(map[string]MigrationInfo),
	}

	for _, dir := range dirs {
		if dir.IsDir() {
			continue
		}

		filename := dir.Name()
		fileExt := strings.ToLower(path.Ext(filename))
		if fileExt != ".sql" {
			continue
		}

		migrationName := filename[:len(filename)-4]
		isRollback := false
		if len(migrationName) > 9 && migrationName[len(migrationName)-9:] == "_rollback" {
			isRollback = true
			migrationName = migrationName[:len(migrationName)-9]
		}

		data, exist := migrations.infos[migrationName]
		if !exist {
			migrations.names = append(migrations.names, migrationName)
			data = MigrationInfo{
				Name: migrationName,
			}
		}

		if isRollback {
			data.DownFilename = path.Join(migrationPath, filename)
		} else {
			data.UpFilename = path.Join(migrationPath, filename)
		}

		migrations.infos[migrationName] = data
	}

	return migrations, nil
}
