package config

import "testing"

func TestMongoConfigUsesEnvironmentValues(t *testing.T) {
	t.Setenv("MONGO_URI", "mongodb://127.0.0.1:27018")
	t.Setenv("MONGO_DATABASE", "portfolio_test")
	if got := MongoURI(); got != "mongodb://127.0.0.1:27018" {
		t.Fatalf("MongoURI() = %q", got)
	}
	if got := MongoDatabaseName(); got != "portfolio_test" {
		t.Fatalf("MongoDatabaseName() = %q", got)
	}
}

func TestMongoConfigUsesLocalDefaults(t *testing.T) {
	t.Setenv("MONGO_URI", "")
	t.Setenv("MONGO_DATABASE", "")
	if got := MongoURI(); got != "mongodb://localhost:27017" {
		t.Fatalf("MongoURI() = %q", got)
	}
	if got := MongoDatabaseName(); got != "portfolio_test" {
		t.Fatalf("MongoDatabaseName() = %q, want portfolio_test", got)
	}
}

func TestValidateTestDatabaseNamesRequiresExplicitIsolatedDatabases(t *testing.T) {
	tests := []struct {
		name       string
		postgresDB string
		mongoDB    string
		wantErr    bool
	}{
		{name: "approved test databases", postgresDB: "pertemuan10_test", mongoDB: "portfolio_test"},
		{name: "main MongoDB database", postgresDB: "pertemuan10_test", mongoDB: "praktikum_backend", wantErr: true},
		{name: "main PostgreSQL database", postgresDB: "praktikum_backend", mongoDB: "portfolio_test", wantErr: true},
		{name: "empty configuration", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTestDatabaseNames(tt.postgresDB, tt.mongoDB)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateTestDatabaseNames() error = %v, wantErr %t", err, tt.wantErr)
			}
		})
	}
}
