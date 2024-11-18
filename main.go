package main

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strconv"
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
	fmt.Println("version: 1.0.0-beta")
	fmt.Println()
}

func echoHelp() {
	fmt.Println(`Commands:`)
	fmt.Println(`  init                Init a Blueprint repo in current work directory`)
	fmt.Println(`  run                 Exec migrations`)
	fmt.Println(`  create, update      Create a pair(include rollback) migration sql files`)
	fmt.Println(`  dump               Dump schema from database`)
	fmt.Println(`  rollback           Rollback`)
	fmt.Println(`                        --step  specify how many step(s) for rollback`)
	fmt.Println(`                        --batch specify how many batch(es) for rollback`)
	fmt.Println(`                        Only one of --step or --batch can be specified at a time,`)
	fmt.Println(`                        default is --batch 1`)
	fmt.Println(`  help                Display this infomation`)
}

var dbs []*sql.DB // 数据库连接

func bootstrap(workDir string) {
	err := loadJsonConfig(workDir)
	if err != nil {
		fmt.Println("Parse config failed:", err.Error())
		os.Exit(1)
	}

	for _, dbCnf := range config.Databases {
		db, err := connectToDb(dbCnf.Host, dbCnf.Port, dbCnf.User, dbCnf.Pass, dbCnf.Name)
		if err != nil {
			fmt.Printf("connect to db[%s] error: %s\n", dbCnf.Host, err)
			os.Exit(1)
		}
		dbs = append(dbs, db)
	}
}

func cleanup() {
	for _, db := range dbs {
		db.Close()
	}
}

func main() {
	echoVersion()

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println("Get cwd failed:", err.Error())
		return
	}

	args := os.Args
	if len(args) == 1 {
		bootstrap(cwd)
		defer cleanup()
		err = runMigration(cwd, dbs)
	} else {
		action := strings.ToLower(args[1])
		params := args[2:]

		switch action {
		case "init":
			err = initBlueprint(cwd)

		case "run":
			bootstrap(cwd)
			defer cleanup()
			err = runMigration(cwd, dbs)

		case "create",
			"update":
			err = createMigration(cwd, action, params)

		case "dump":
			bootstrap(cwd)
			defer cleanup()
			forceDump := false
			for _, param := range params {
				if param == "--force" {
					forceDump = true
				}
			}
			err = dumpSchemas(dbs[0], cwd, forceDump)

		case "rollback":
			bootstrap(cwd)
			defer cleanup()
			step := 0
			batch := 0
			for idx := 0; idx < len(params); idx++ {
				param := params[idx]
				if param == "--step" || param == "--batch" {
					if idx+1 >= len(params) {
						err = errors.New("invalid param: " + param)
						break
					}
					value := 0
					value, err = strconv.Atoi(params[idx+1])
					if err != nil {
						break
					}
					if value < 1 {
						err = errors.New("invalid param value: " + param + " = " + params[idx+1])
						break
					}
					switch param {
					case "--step":
						step = value
					case "--batch":
						batch = value
					}
					idx++
				}
			}
			if step != 0 && batch != 0 {
				err = errors.New("only one of --step or --batch can be specified at a time")
			}
			if err == nil {
				err = rollbackMigration(cwd, dbs, step, batch)
			}

		case "help":
			fallthrough
		default:
			echoHelp()
			err = nil
		}
	}

	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
