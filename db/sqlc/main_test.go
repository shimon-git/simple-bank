package db

import (
	"database/sql"
	"log"
	"os"
	"testing"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

const (
	envFile  = "../../env/.db.env"
	dbDriver = "postgres"
)

var testQueries *Queries
var testDB *sql.DB

func loadSecrets() {
	if err := godotenv.Load(envFile); err != nil {
		log.Fatalf("Error loading %s: %v", envFile, err)
	}
}

func TestMain(m *testing.M) {
	loadSecrets()
	var err error
	testDB, err = sql.Open(dbDriver, os.Getenv("CONNECTION_STRING"))
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	testQueries = New(testDB)

	os.Exit(m.Run())
}
