package db

import (
	"database/sql"
	"log"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"github.com/shimon-git/simple-bank/util"
)

const (
	confFolder = "../.."
)

var testQueries *Queries
var testDB *sql.DB

func TestMain(m *testing.M) {
	// load the config file from the given folder
	config, err := util.LoadConfig(confFolder)
	if err != nil {
		log.Fatalf("failed to load configurations: %v", err)
	}

	testDB, err = sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	testQueries = New(testDB)

	os.Exit(m.Run())
}
