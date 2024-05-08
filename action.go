package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
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
		host, _ := input(fmt.Sprintf("Input the host of DB[%d] (default: 127.0.0.1): ", i))
		if host == "" {
			host = "127.0.0.1"
		}
		portStr, _ := input(fmt.Sprintf("Input the port of DB[%d] (default: 3306): ", i))
		port, _ := strconv.Atoi(portStr)
		if port == 0 {
			port = 3306
		}
		user, _ := input(fmt.Sprintf("Input the user of DB[%d]: ", i))
		pass, _ := input(fmt.Sprintf("Input the pass of DB[%d]: ", i))
		name, _ := input(fmt.Sprintf("Input the name of DB[%d]: ", i))
		cnf.Databases = append(cnf.Databases, MySQLConfig{
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

func runMigration(migrationPath string, dbs []*sql.DB) error {
	// 读取 migration 文件
	migrations, err := LoadMigrations(migrationPath)
	if err != nil {
		return err
	}

	for _, db := range dbs {
		err := checkMigraionInfoTable(db)
		if err != nil {
			return fmt.Errorf("check migration info failed: %s", err.Error())
		}

		maxBatch := uint(0)
		recs, err := getMigrationInfos(db)
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
		err = DoTransaction(db, func(tx *sql.Tx) error {
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
				err = execMigration(tx, upSQL)
				if err != nil {
					return err
				}
				err = insertMigrationInfo(tx, MigrationRec{
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

func getMigrationFilename(note string) (name, rollbackName string) {
	now := time.Now()
	name = fmt.Sprintf("%s_%s.sql", now.Format("200601021504"), note)
	rollbackName = fmt.Sprintf("%s_%s_rollback.sql", now.Format("200601021504"), note)
	return
}
