package main

import (
	"database/sql"
	"fmt"
	"os"
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

func echoHelp() {
	//
}

var db *sql.DB // 数据库连接
func bootstrap(workDir string) {
	err := loadJsonConfig(workDir)
	if err != nil {
		fmt.Println("Parse config failed:", err.Error())
		os.Exit(1)
	}

	db, err = connectToDb(config.Host, config.Port, config.User, config.Pass, config.Name)
	if err != nil {
		fmt.Println("connect to db error:", err)
		os.Exit(1)
	}
}

func cleanup() {
	db.Close()
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
		runMigration(cwd, db)
	} else {
		action := strings.ToLower(args[1])
		params := args[2:]

		switch action {
		case "init":
			err = initBlueprint(cwd)

		case "run":
			bootstrap(cwd)
			defer cleanup()
			runMigration(cwd, db)

		case "create",
			"update":
			err = createMigration(cwd, action, params)

		default:
			echoHelp()
			err = nil
		}

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		return
	}
}
