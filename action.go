package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const BlueprintConfigFileName string = "blueprint.json"

var reader *bufio.Reader

func init() {
	reader = bufio.NewReader(os.Stdin)
}

func initBlueprint(workDir string) error {
	isBlueprintRepo, err := isBlueprintRepo(workDir)
	if err != nil {
		return err
	}
	if isBlueprintRepo {
		return errors.New("Reinitialized existing Blueprint repository in " + path.Join(workDir, "blueprint.json"))
	}

	cnf := Config{}

	env, err := input("Environment(local/test/production, default: local): ")
	if err != nil {
		return err
	}
	if env == "" {
		env = "local"
	}
	cnf.Env = env

	i := 1
	for {
		dbType, _ := input(fmt.Sprintf("Input the type of DB[%d] (mysql/pg, default: mysql): ", i))
		if dbType == "" {
			dbType = "mysql"
		}
		host, _ := input(fmt.Sprintf("Input the host of DB[%d] (default: 127.0.0.1): ", i))
		if host == "" {
			host = "127.0.0.1"
		}
		defaultPort := 3306
		if dbType == "pg" {
			defaultPort = 5432
		}
		portStr, _ := input(fmt.Sprintf("Input the port of DB[%d] (default: %d): ", i, defaultPort))
		port, _ := strconv.Atoi(portStr)
		if port == 0 {
			port = defaultPort
		}
		user, _ := input(fmt.Sprintf("Input the user of DB[%d]: ", i))
		pass, _ := input(fmt.Sprintf("Input the pass of DB[%d]: ", i))
		name, _ := input(fmt.Sprintf("Input the name of DB[%d]: ", i))
		cnf.Databases = append(cnf.Databases, DBConfig{
			Type: dbType,
			Host: host,
			Port: uint(port),
			User: user,
			Pass: pass,
			Name: name,
		})

		more, _ := input("Add another db info? (yN): ")
		if strings.ToLower(more) == "y" || strings.ToLower(more) == "yes" {
			i++
		} else {
			break
		}
	}

	cnfBytes, err := json.MarshalIndent(cnf, "", "    ")
	if err != nil {
		return err
	}

	err = os.WriteFile(
		path.Join(workDir, BlueprintConfigFileName),
		cnfBytes,
		0644)
	if err != nil {
		return err
	}

	fmt.Println("Initialized Blueprint repository in " + path.Join(workDir, BlueprintConfigFileName))
	return nil
}

func runMigration(migrationPath string, dbs []*DBConnection) error {
	// 读取 migration 文件
	migrations, err := LoadMigrations(migrationPath)
	if err != nil {
		return err
	}

	for _, db := range dbs {
		err := db.Driver.CheckMigrationInfoTable(db.DB)
		if err != nil {
			return fmt.Errorf("check migration info failed: %s", err.Error())
		}

		maxBatch := uint(0)
		recs, err := db.Driver.GetMigrationInfos(db.DB)
		if err != nil {
			return fmt.Errorf("get migration infos error: %s", err.Error())
		}
		recMap := make(map[string]struct{})
		for _, rec := range recs {
			if rec.Batch > maxBatch {
				maxBatch = rec.Batch
			}
			recMap[rec.Migration] = struct{}{}
		}

		maxBatch++
		err = DoTransaction(db.DB, func(tx *sql.Tx) error {
			for idx, name := range migrations.GetNames() {
				if _, exist := recMap[name]; exist {
					fmt.Printf("[%d] %s had excuted, skip\n", idx, name)
					continue
				}
				fmt.Printf("[%d] %s\n", idx, name)
				migration := migrations.GetInfo(name)
				err := migration.LoadSQLFile()
				if err != nil {
					return err
				}
				upSQL := migration.upSQL
				err = db.Driver.ExecMigration(tx, upSQL)
				if err != nil {
					return err
				}
				err = db.Driver.InsertMigrationInfo(tx, MigrationRec{
					Migration: name,
					Batch:     maxBatch,
				})
				if err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// 创建一对 Migration 文件
func createMigration(workDir, action string, params []string) error {
	if ok, _ := isBlueprintRepo(workDir); !ok {
		return errors.New("Not a Blueprint repository (" + workDir + ")")
	}

	tableName := ""
	if len(params) == 0 {
		tableName, _ = input("input table name: ")
	} else {
		tableName = params[0]
	}
	name, rbName := getMigrationFilename(fmt.Sprintf("%s_%s", action, tableName))
	f, err := os.Create(path.Join(workDir, name))
	if err != nil {
		return errors.New("Create migration file failed: " + err.Error())
	}
	f.Close()

	f, err = os.Create(path.Join(workDir, rbName))
	if err != nil {
		return errors.New("Create migration rollback file failed: " + err.Error())
	}
	f.Close()

	return nil
}

// 导出表结构
func dumpSchemas(db *DBConnection, workDir string, forceDump bool) error {
	repoEmpty, err := isEmptyRepo(workDir)
	if err != nil {
		return err
	}

	if !repoEmpty && !forceDump {
		return fmt.Errorf("it seems %s is not a empty repository, use --force to dump anyway", workDir)
	}

	tables, err := db.Driver.GetTables(db.DB)
	if err != nil {
		return err
	}

	// Filter out migrations table
	filteredTables := make([]string, 0)
	for _, table := range tables {
		if table == "migrations" {
			continue
		}
		filteredTables = append(filteredTables, table)
	}
	tables = filteredTables

	creations := make(map[string]string)
	for _, table := range tables {
		creation, err := db.Driver.ShowTableCreate(db.DB, table)
		if err != nil {
			return err
		}
		creations[table] = creation
	}

	writeSqlFile := func(table, creationFilename, creation, rollbackFilename, rollback string) error {
		fmt.Printf("Writing table[%s] creation to %s\n", table, creationFilename)
		err := os.WriteFile(
			creationFilename,
			[]byte(creation),
			0644,
		)
		if err != nil {
			return err
		}

		fmt.Printf("Writing table[%s] rollback to %s\n", table, rollbackFilename)
		err = os.WriteFile(
			rollbackFilename,
			[]byte(rollback),
			0644,
		)
		return err
	}

	for _, table := range tables {
		mName, rbName := getMigrationFilename(fmt.Sprintf("create_%s", strings.ToLower(table)))
		creation := creations[table]
		rollback := fmt.Sprintf("DROP TABLE `%s`;\n", table)

		err = writeSqlFile(
			table,
			path.Join(workDir, mName), creation,
			path.Join(workDir, rbName), rollback,
		)
		if err != nil {
			return err
		}
	}

	return nil
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

// 回滚
func rollbackMigration(migrationPath string, dbs []*DBConnection, step, batch int) error {
	migrations, err := LoadMigrations(migrationPath)
	if err != nil {
		return err
	}

	if step == 0 && batch == 0 {
		batch = 1
	}

	for _, db := range dbs {
		recs, err := db.Driver.GetMigrationInfos(db.DB)
		if err != nil {
			return err
		}

		if len(recs) == 0 {
			return fmt.Errorf("nothing to rollback")
		}

		remainStep := 0
		byStep := false
		if step > 0 {
			remainStep = step
			byStep = true
		}

		remainBatch := 0
		lastBatch := uint(0)
		byBatch := false
		if batch > 0 {
			remainBatch = batch
			byBatch = true
		}

		list := make([]MigrationRec, 0)
		for i := len(recs) - 1; i >= 0; i-- {
			if byStep {
				if remainStep == 0 {
					break
				}
				list = append(list, recs[i])
				remainStep--
				continue
			}

			if byBatch {
				if lastBatch == 0 {
					lastBatch = recs[i].Batch
				}

				if lastBatch != recs[i].Batch {
					remainBatch--
				}

				if remainBatch == 0 {
					break
				}

				list = append(list, recs[i])
				lastBatch = recs[i].Batch
				continue
			}
		}

		err = DoTransaction(db.DB, func(tx *sql.Tx) error {
			for _, migrRec := range list {
				migration := migrations.GetInfo(migrRec.Migration)

				// 执行回滚
				err = migration.LoadSQLFile()
				if err != nil {
					return err
				}
				downSQL := migration.GetDownSQL()
				err = db.Driver.ExecMigration(tx, downSQL)
				if err != nil {
					return err
				}

				// 删除 migration 记录
				err = db.Driver.DeleteMigrationInfo(tx, migrRec.Id)
				if err != nil {
					return err
				}
				fmt.Printf("[%d] Batch[%d] %s rolled back\n", migrRec.Id, migrRec.Batch, migrRec.Migration)
			}
			return nil
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func input(prompt string) (string, error) {
	fmt.Print(prompt)
	input, err := reader.ReadString('\n')
	return strings.Trim(input, " \n"), err
}

func isBlueprintRepo(workDir string) (bool, error) {
	_, err := os.Stat(path.Join(workDir, BlueprintConfigFileName))
	if err == nil {
		return true, nil
	}

	if !os.IsNotExist(err) {
		return false, err
	}
	return false, nil
}

func isEmptyRepo(workDir string) (bool, error) {
	// 检查传入的路径是否为目录
	fileInfo, err := os.Stat(workDir)
	if err != nil {
		return false, err
	}
	if !fileInfo.IsDir() {
		return false, fmt.Errorf("%s is not a dir", workDir)
	}

	// 读取目录下的文件
	entries, err := os.ReadDir(workDir)
	if err != nil {
		return false, err
	}

	// 检查目录下是否有 .sql 文件
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".sql" {
			return false, nil
		}
	}

	return true, nil
}

func getMigrationFilename(note string) (name, rollbackName string) {
	now := time.Now()
	name = fmt.Sprintf("%s_%s.sql", now.Format("200601021504"), note)
	rollbackName = fmt.Sprintf("%s_%s_rollback.sql", now.Format("200601021504"), note)
	return
}
