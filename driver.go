package main

import (
	"database/sql"
)

type DatabaseDriver interface {
	Connect(host string, port uint, user, pass, dbName string) (*sql.DB, error)
	CheckMigrationInfoTable(db *sql.DB) error
	GetMigrationInfos(db *sql.DB) ([]MigrationRec, error)
	InsertMigrationInfo(db *sql.Tx, info MigrationRec) error
	DeleteMigrationInfo(db *sql.Tx, id uint) error
	ExecMigration(db *sql.Tx, migrationSQL string) error
	ShowTableCreate(db *sql.DB, table string) (string, error)
	GetTables(db *sql.DB) ([]string, error)
}

type DBConnection struct {
	*sql.DB
	Driver DatabaseDriver
}
