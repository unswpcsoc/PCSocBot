package main

import (
	"database/sql"
	//"encoding/json"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"

	"github.com/unswpcsoc/PCSocBot/commands"
)

// use tag implementation
type tag struct {
	UID      string
	Username string // don't trust this, always fetch from the UID
	Tag      string
	Platform string
	PingMe   bool
}

type platform struct {
	Name  string
	Role  interface{}
	Users map[string]*tag // indexed by user id's
}

// TODO: default games and api integrations
type tagStorer struct {
	Platforms map[string]*platform
}

func (t *tagStorer) Index() string { return "tags" }

func main() {
	var err error

	// open db to migrate
	db, err := sql.Open("sqlite3", "db.sqlite3")
	if err != nil {
		log.Fatalln(err)
	}
	defer db.Close()

	// open db to create
	err = commands.DBOpen("bot.db")
	defer commands.DBClose()

	fmt.Println("Opened both dbs...")

	rows, err := db.Query("select * from tag")
	if err != nil {
		log.Fatalln(err)
	}
	defer rows.Close()

	/*
		cols, _ := rows.Columns()
		for _, str := range cols {
			fmt.Println(str)
		}
	*/

	// create tagStorer
	tgs := &tagStorer{
		Platforms: make(map[string]*platform),
	}

	fmt.Println("Iterating rows...")

	for rows.Next() {
		var uid, plt, tagg string
		rows.Scan(&uid, &plt, &tagg)

		// check if platform already exists
		plat, ok := tgs.Platforms[plt]
		if !ok {
			// new platform, make it and add it to the db
			plat = &platform{
				Name:  plt,
				Role:  nil,
				Users: make(map[string]*tag),
			}
			tgs.Platforms[plt] = plat
		}

		// create tag and enter it into the platform
		plat.Users[uid] = &tag{
			UID:      uid,
			Username: "",
			Tag:      tagg,
			Platform: plt,
			PingMe:   false,
		}
	}

	/*
		b, err := json.MarshalIndent(tgs, "", "    ")
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Println(string(b))
	*/

	// commit to the db
	commands.DBSet(tgs, "fulltags")

	fmt.Println("Committed changes")
}
