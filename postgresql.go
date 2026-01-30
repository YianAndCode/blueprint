package main

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	_ "github.com/lib/pq"
)

type PostgreSQLDriver struct{}

func (d PostgreSQLDriver) Connect(host string, port uint, user, pass, dbName string) (*sql.DB, error) {
	// sslmode=disable is used for simplicity in development environments.
	// connect_timeout is set to 5 seconds to avoid long waits on unreachable servers.
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable&connect_timeout=5", user, pass, host, port, dbName)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func (d PostgreSQLDriver) CheckMigrationInfoTable(db *sql.DB) error {
	// Check if table exists
	var exists bool
	query := "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'migrations')"
	err := db.QueryRow(query).Scan(&exists)
	if err != nil {
		return err
	}

	if !exists {
		_, err := db.Exec(`
			CREATE TABLE migrations (
			  id SERIAL PRIMARY KEY,
			  migration VARCHAR(255) NOT NULL,
			  batch INTEGER NOT NULL
			);
		`)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d PostgreSQLDriver) GetMigrationInfos(db *sql.DB) ([]MigrationRec, error) {
	info := make([]MigrationRec, 0)

	rows, err := db.Query("SELECT id, migration, batch FROM migrations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		rec := MigrationRec{}
		err = rows.Scan(&rec.Id, &rec.Migration, &rec.Batch)
		if err != nil {
			return nil, err
		}
		info = append(info, rec)
	}

	return info, nil
}

func (d PostgreSQLDriver) InsertMigrationInfo(db *sql.Tx, info MigrationRec) error {
	_, err := db.Exec(`
		INSERT INTO migrations (migration, batch)
		VALUES ($1, $2);
	`, info.Migration, info.Batch)
	if err != nil {
		return err
	}

	return nil
}

func (d PostgreSQLDriver) DeleteMigrationInfo(db *sql.Tx, id uint) error {
	_, err := db.Exec(`DELETE FROM migrations WHERE id = $1`, id)
	return err
}

func (d PostgreSQLDriver) ExecMigration(db *sql.Tx, migrationSQL string) error {
	// Basic split by semicolon.
	// Note: This might be fragile for complex PG statements involving $$ quoting.
	statements := strings.Split(migrationSQL, ";")

	for _, statement := range statements {
		statement = strings.TrimSpace(statement)
		if statement == "" {
			continue
		}

		_, err := db.Exec(statement)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d PostgreSQLDriver) ShowTableCreate(db *sql.DB, table string) (string, error) {
	// PostgreSQL does not have a native SHOW CREATE TABLE command.
	// Implementing a full schema dump is complex.
	// Returning a placeholder to indicate it's not fully supported in this simple implementation.
	return fmt.Sprintf("-- SHOW CREATE TABLE is not supported for PostgreSQL in this tool yet.\n-- Please use pg_dump for schema exports of table: %s", table),
		errors.New("Dump table is not supported for PostgreSQL in this tool yet, please use pg_dump instead")
}

func (d PostgreSQLDriver) GetTables(db *sql.DB) ([]string, error) {
	tables := make([]string, 0)
	rows, err := db.Query("SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' AND table_type = 'BASE TABLE'")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		table := ""
		err = rows.Scan(&table)
		if err != nil {
			return nil, err
		}
		tables = append(tables, table)
	}
	return tables, nil
}
