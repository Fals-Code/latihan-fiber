package config

import (
	"fmt"
	"os"
)

const (
	PostgreSQLTestDatabase = "pertemuan10_test"
	MongoTestDatabase      = "portfolio_test"
)

func MongoURI() string {
	if value := os.Getenv("MONGO_URI"); value != "" {
		return value
	}
	return "mongodb://localhost:27017"
}

func MongoDatabaseName() string {
	if value := os.Getenv("MONGO_DATABASE"); value != "" {
		return value
	}
	return MongoTestDatabase
}

// ValidateTestDatabaseNames prevents database tests from using project databases.
func ValidateTestDatabaseNames(postgresDatabase, mongoDatabase string) error {
	if postgresDatabase != PostgreSQLTestDatabase {
		return fmt.Errorf("DB_NAME must be %q for database tests", PostgreSQLTestDatabase)
	}
	if mongoDatabase != MongoTestDatabase {
		return fmt.Errorf("MONGO_DATABASE must be %q for database tests", MongoTestDatabase)
	}
	return nil
}
