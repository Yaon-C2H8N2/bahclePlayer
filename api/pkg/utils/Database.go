package utils

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

var Database *pgxpool.Pool

func InitDatabase() {
	connectionString := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_URL"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_DB"),
	)

	pool, err := pgxpool.New(context.Background(), connectionString)
	if err != nil {
		panic(err)
	}

	Database = pool
	migrateDatabase(pool)
}

func GetConnection() *pgxpool.Conn {
	ctx, _ := context.WithDeadline(context.Background(), time.Now().Add(20*time.Second))
	conn, err := Database.Acquire(ctx)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
	}

	return conn
}

func DoRequest(conn *pgxpool.Conn, query string, args ...any) pgx.Rows {
	ctx, _ := context.WithDeadline(context.Background(), time.Now().Add(10*time.Second))

	rows, err := conn.Query(ctx, query, args...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Query failed: %v\n", err)
	}

	return rows
}

func migrateDatabase(pool *pgxpool.Pool) {
	db := stdlib.OpenDBFromPool(pool)
	fmt.Printf("Migrating database %s\n", pool.Config().ConnConfig.ConnString())

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
