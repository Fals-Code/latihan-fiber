package database

import (
	"context"
	"testing"
)

func TestNewMongoClientRejectsUnsupportedURI(t *testing.T) {
	client, err := NewMongoClient(context.Background(), "http://localhost:27017")
	if err == nil {
		if client != nil {
			_ = client.Disconnect(context.Background())
		}
		t.Fatal("expected an unsupported MongoDB URI scheme to be rejected")
	}
}
