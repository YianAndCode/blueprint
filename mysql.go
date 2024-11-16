package main

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

// 连接到数据库
func connectToDb(host string, port uint, user, pass, dbName string) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", user, pass, host, port, dbName)
	return sql.Open("mysql", dsn)
}

// 检查表是否存在
func checkMigraionInfoTable(db *sql.DB) error {
	rows, err := db.Query("SHOW TABLES LIKE 'migrations'")
	if err != nil {
		return err
	}
	defer rows.Close()

	if !rows.Next() {
		_, err := db.Exec(`
			CREATE TABLE migrations (
			  id int unsigned NOT NULL AUTO_INCREMENT,
			  migration varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
			  batch int unsigned NOT NULL,
			  PRIMARY KEY (id)
			) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
		`)
		if err != nil {
			return err
		}
	}

	return nil
}

// 获取过往 migration 记录
func getMigrationInfos(db *sql.DB) ([]MigrationRec, error) {
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

// 插入 migration 记录
func insertMigrationInfo(db *sql.Tx, info MigrationRec) error {
	_, err := db.Exec(`
		INSERT INTO migrations (migration, batch)
		VALUES (?, ?);
	`, info.Migration, info.Batch)
	if err != nil {
		return err
	}

	return nil
}

// 删除 migration 记录
func deleteMigrationInfo(db *sql.Tx, id uint) error {
	_, err := db.Exec(`DELETE FROM migrations WHERE id = ?`, id)
	return err
}

// 执行迁移
func execMigration(db *sql.Tx, migrationSQL string) error {
	// 分割多条语句
	statements := strings.Split(migrationSQL, ";")

	// 逐个执行每条语句
	for _, statement := range statements {
		statement = strings.TrimSpace(statement)
		if statement == "" {
			// 跳过空语句
			continue
		}

		// 执行语句
		_, err := db.Exec(statement)
		if err != nil {
			return err
		}
	}

	return nil
}

// 指定表结构
func showTableCreate(db *sql.DB, table string) (string, error) {
	rows, err := db.Query("SHOW CREATE TABLE " + table)
	if err != nil {
		return "", err
	}
	creation := ""
	rows.Next()
	err = rows.Scan(&table, &creation)
	if err != nil {
		return "", err
	}

	return creation, nil
}

// 事务
func DoTransaction(db *sql.DB, fn func(tx *sql.Tx) error) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	err = fn(tx)
	if err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}
