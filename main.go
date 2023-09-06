package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/shimon-git/simple-bank/api"
	db "github.com/shimon-git/simple-bank/db/sqlc"
)

const (
	envFile       = "./env/.db.env"
	dbDriver      = "postgres"
	serverAddress = "0.0.0.0:80"
)

func loadSecrets() {
	if err := godotenv.Load(envFile); err != nil {
		log.Fatalf("Error loading %s: %v", envFile, err)
	}
}

func main() {
	loadSecrets()
	conn, err := sql.Open(dbDriver, os.Getenv("CONNECTION_STRING"))
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	store := db.NewStore(conn)

	server := api.NewServer(store)

	if err = server.Start(serverAddress); err != nil {
		log.Fatal("cannot starting the server on:", serverAddress, ":", err)
	}
}
