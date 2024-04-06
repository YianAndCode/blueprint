package main

import (
	"database/sql"
	"fmt"
	"os"
	"path"
	"strings"
)

func echoVersion() {
	fmt.Println(`    ____     __                                   _            __ `)
	fmt.Println(`   / __ )   / /  __  __  ___     ____    _____   (_)   ____   / /_`)
	fmt.Println(`  / __  |  / /  / / / / / _ \   / __ \  / ___/  / /   / __ \ / __/`)
	fmt.Println(` / /_/ /  / /  / /_/ / /  __/  / /_/ / / /     / /   / / / // /_  `)
	fmt.Println(`/_____/  /_/   \__,_/  \___/  / .___/ /_/     /_/   /_/ /_/ \__/  `)
	fmt.Println(`                             /_/                                  `)
	fmt.Println()
	fmt.Println("version: 0.0.1")
	fmt.Println()
}

func main() {
	echoVersion()

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println("Get cwd failed:", err.Error())
		return
	}

	err = loadJsonConfig(cwd)
	if err != nil {
		fmt.Println("Parse config failed:", err.Error())
		return
	}

	db, err := connectToDb(config.Host, config.Port, config.User, config.Pass, config.Name)
	if err != nil {
		fmt.Println("connect to db error:", err)
		return
	}
	defer db.Close()

	err = checkMigraionInfoTable(db)
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

	migrationPath := cwd

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
