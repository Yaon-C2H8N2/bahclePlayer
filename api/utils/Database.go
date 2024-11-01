package utils

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"
	"log"
	"os"
)

func GetConnection() *pgx.Conn {
	connectionString := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_URL"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_DB"),
	)

	conn, err := pgx.Connect(context.Background(), connectionString)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
	}

	return conn
}

func DoRequest(conn *pgx.Conn, query string, args ...any) pgx.Rows {
	rows, err := conn.Query(context.Background(), query, args...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Query failed: %v\n", err)
	}

	return rows
}

func Migrate() {
	connectionString := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_URL"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_DB"),
	)

	fmt.Println("Migrating database...")
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Fatal(err)
		}
	}()
	instance, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatal(err)
	}
	m, err := migrate.NewWithDatabaseInstance("file://./sql/", os.Getenv("POSTGRES_DB"), instance)
	if err != nil {
		log.Fatal(err)
	}
	if err := m.Up(); err != nil {
		if err.Error() == "no change" {
			log.Println("No changes to migrate.")
		} else {
			log.Fatal(err)
		}
	}
	fmt.Println("Migration complete.")
}
