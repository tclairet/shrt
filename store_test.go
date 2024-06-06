package main

import (
	"context"
	"os"
	"testing"
)

func TestPostgres(t *testing.T) {
	tests := []struct {
		name  string
		db    string
		short string
		long  string
	}{
		{
			name:  "postgres",
			db:    "mem",
			short: "foo",
			long:  "bar",
		},
		{
			name:  "mem",
			db:    "postgres",
			short: "foo",
			long:  "bar",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := storeFactory(t, tt.db)
			if err := db.Save(tt.short, tt.long); err != nil {
				t.Fatal(err)
			}
			long, err := db.Long(tt.short)
			if err != nil {
				t.Fatal(err)
			}
			if got, want := long, tt.long; got != want {
				t.Errorf("got %v want %v", got, want)
			}
			exist, err := db.Exist(tt.short)
			if err != nil {
				t.Fatal(err)
			}
			if !exist {
				t.Errorf("should exist")
			}
			exist, err = db.Exist("foobar")
			if err != nil {
				t.Fatal(err)
			}
			if exist {
				t.Errorf("should not exist")
			}
		})
	}
}

func storeFactory(t *testing.T, storeType string) store {
	switch storeType {
	case "mem":

		return newMem()
	case "postgres":
		return newPostgresTest(t)
	}
	t.Fatalf("unknown store type %s", storeType)
	return nil
}

func newPostgresTest(t *testing.T) store {
	t.Helper()
	if _, exists := os.LookupEnv("DB_URL"); !exists {
		t.Skip("DB_URL not configured")
	}
	db, err := newPostgres(os.Getenv("DB_URL"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, err := db.pool.Exec(context.Background(), "TRUNCATE urls CASCADE"); err != nil {
			t.Fatal(err)
		}
	})
	return db
}
