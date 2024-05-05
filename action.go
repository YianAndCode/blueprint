package main

import (
	"bufio"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
	"time"
)

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

	// TODO:

	return nil
}

func runMigration(workDir string, db *sql.DB) {
	err := checkMigraionInfoTable(db)
	if err != nil {
		fmt.Println("check migration info failed:", err)
		return
	}

	maxBatch := uint(0)
	recs, err := getMigrationInfos(db)
	if err != nil {
		fmt.Println("get migration infos error:", err)
		return
	}
	recMap := make(map[string]struct{})
	for _, rec := range recs {
		if rec.Batch > maxBatch {
			maxBatch = rec.Batch
		}
		recMap[rec.Migration] = struct{}{}
	}

	migrationPath := workDir

	dirs, err := os.ReadDir(migrationPath)
	if err != nil {
		fmt.Println(err)
	}

	migrationNames := make([]string, 0)
	migrations := make(map[string]Migration)

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

		migr, exist := migrations[migrationName]
		if !exist {
			migrationNames = append(migrationNames, migrationName)
			migr = Migration{
				Name: migrationName,
			}
		}

		if isRollback {
			migr.DownFilename = path.Join(migrationPath, filename)
		} else {
			migr.UpFilename = path.Join(migrationPath, filename)
		}

		migrations[migrationName] = migr
	}

	maxBatch++
	err = DoTransaction(db, func(tx *sql.Tx) error {
		for idx, name := range migrationNames {
			if _, exist := recMap[name]; exist {
				fmt.Printf("[%d] %s had excuted, skip\n", idx, name)
				continue
			}
			fmt.Printf("[%d] %s\n", idx, name)
			migration := migrations[name]
			err := migration.LoadSQLFile()
			if err != nil {
				return err
			}
			upSQL := migration.upSQL
			err = execMigration(tx, upSQL)
			if err != nil {
				return err
			}
			err = insertMigrationInfo(tx, MigrationInfo{
				Migration: name,
				Batch:     maxBatch,
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	fmt.Println(err)
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
	_, err := os.Stat(path.Join(workDir, "blueprint.json"))
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
