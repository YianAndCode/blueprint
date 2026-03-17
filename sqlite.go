package main

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type SQLiteDriver struct{}

func (d SQLiteDriver) Connect(host string, port uint, user, pass, dbName string) (*sql.DB, error) {
	// For sqlite, dbName contains the file path (handled in main.go)
	db, err := sql.Open("sqlite3", dbName)
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

func (d SQLiteDriver) CheckMigrationInfoTable(db *sql.DB) error {
	// Check if table exists
	query := "SELECT name FROM sqlite_master WHERE type='table' AND name='migrations'"
	rows, err := db.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	if !rows.Next() {
		_, err := db.Exec(`
			CREATE TABLE migrations (
			  id INTEGER PRIMARY KEY AUTOINCREMENT,
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

func (d SQLiteDriver) GetMigrationInfos(db *sql.DB) ([]MigrationRec, error) {
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

func (d SQLiteDriver) InsertMigrationInfo(db *sql.Tx, info MigrationRec) error {
	_, err := db.Exec(`
		INSERT INTO migrations (migration, batch)
		VALUES (?, ?);
	`, info.Migration, info.Batch)
	if err != nil {
		return err
	}

	return nil
}

func (d SQLiteDriver) DeleteMigrationInfo(db *sql.Tx, id uint) error {
	_, err := db.Exec(`DELETE FROM migrations WHERE id = ?`, id)
	return err
}

func (d SQLiteDriver) ExecMigration(db *sql.Tx, migrationSQL string) error {
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

func (d SQLiteDriver) ShowTableCreate(db *sql.DB, table string) (string, error) {
	query := fmt.Sprintf("SELECT sql FROM sqlite_master WHERE type='table' AND name='%s'", table)
	rows, err := db.Query(query)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	
	if rows.Next() {
		var creation sql.NullString
		err = rows.Scan(&creation)
		if err != nil {
			return "", err
		}
		if creation.Valid {
			return creation.String, nil
		}
	}

	return "", errors.New("table not found or no creation sql available")
}

func (d SQLiteDriver) GetTables(db *sql.DB) ([]string, error) {
	tables := make([]string, 0)
	rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'")
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