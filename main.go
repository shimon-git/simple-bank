package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
	"github.com/shimon-git/simple-bank/api"
	db "github.com/shimon-git/simple-bank/db/sqlc"
	"github.com/shimon-git/simple-bank/util"
)

const (
	confFolder = "."
)

func main() {
	// load the config file from the given folder
	config, err := util.LoadConfig(confFolder)
	if err != nil {
		log.Fatalf("failed to load configurations: %v", err)
	}
	// Creating a DB connection objects
	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}
	// creating a new store object
	store := db.NewStore(conn)
	// creating a new server object
	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatalf("cannot create server: %w", err)
	}
	// starting the server on the given interface and port
	if err = server.Start(config.ServerAddress); err != nil {
		log.Fatal("cannot starting the server on:", config.ServerAddress, ":", err)
	}
}
