// +build ignore

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/soumabali/vexa/internal/db"
)

func main() {
	migrationsDir := flag.String("migrations", "./internal/db/migrations", "Directory containing migration SQL files")
	seedsDir := flag.String("seeds", "", "Directory containing seed SQL files (defaults to <migrations>/../seeds)")
	dsn := flag.String("dsn", os.Getenv("DATABASE_URL"), "PostgreSQL DSN")
	command := flag.String("cmd", "status", "Command: status | up | down | down-to | seed | reset")
	targetVersion := flag.String("version", "", "Target version for rollback commands")
	maxOpenConns := flag.Int("max-open-conns", 25, "Max open connections")
	maxIdleConns := flag.Int("max-idle-conns", 10, "Max idle connections")
	flag.Parse()

	if *dsn == "" {
		fmt.Println("Error: DATABASE_URL environment variable or -dsn flag required")
		os.Exit(1)
	}

	// Default seeds dir to sibling of migrations dir
	if *seedsDir == "" {
		*migrationsDir = strings.TrimRight(*migrationsDir, "/")
		*migrationsDir = strings.TrimRight(*migrationsDir, "/")
		parent := filepath.Dir(*migrationsDir)
		*seedsDir = filepath.Join(parent, "seeds")
	}

	database, err := db.New(*dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Override connection pool settings
	database.SetMaxOpenConns(*maxOpenConns)
	database.SetMaxIdleConns(*maxIdleConns)

	migrator := db.NewMigrator(database, *migrationsDir)

	switch *command {
	case "status":
		migrations, err := migrator.Status()
		if err != nil {
			log.Fatalf("Failed to get migration status: %v", err)
		}
		fmt.Println("=== Migration Status ===")
		if len(migrations) == 0 {
			fmt.Println("No migrations found in directory")
			os.Exit(0)
		}
		for _, m := range migrations {
			status := "PENDING"
			if m.Applied {
				status = "APPLIED"
			}
			fmt.Printf("  [%s] %s - %s\n", status, m.Version, m.Description)
		}

	case "up":
		if err := migrator.EnsureSchemaMigrations(); err != nil {
			log.Fatalf("Failed to init schema_migrations table: %v", err)
		}
		if err := migrator.RunAll(); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		fmt.Println("All migrations applied successfully")

	case "down":
		if err := migrator.EnsureSchemaMigrations(); err != nil {
			log.Fatalf("Failed to init schema_migrations table: %v", err)
		}
		if err := migrator.Rollback(); err != nil {
			log.Fatalf("Rollback failed: %v", err)
		}
		fmt.Println("Rollback completed successfully")

	case "down-to":
		if *targetVersion == "" {
			log.Fatal("Error: -version flag required for down-to command")
		}
		if err := migrator.EnsureSchemaMigrations(); err != nil {
			log.Fatalf("Failed to init schema_migrations table: %v", err)
		}
		if err := migrator.RollbackTo(*targetVersion); err != nil {
			log.Fatalf("Rollback-to failed: %v", err)
		}
		fmt.Println("Rollback-to completed successfully")

	case "seed":
		if err := migrator.Seed(); err != nil {
			log.Fatalf("Seed failed: %v", err)
		}
		fmt.Println("Seeding completed successfully")

	case "reset":
		if err := migrator.EnsureSchemaMigrations(); err != nil {
			log.Fatalf("Failed to init schema_migrations table: %v", err)
		}
		if err := migrator.Reset(); err != nil {
			log.Fatalf("Reset failed: %v", err)
		}
		fmt.Println("Reset completed — all migrations reapplied")

	default:
		fmt.Printf("Unknown command: %s\n", *command)
		fmt.Println("Usage: migrate -cmd=status|up|down|down-to|seed|reset [-version=V] [-dsn=URL]")
		os.Exit(1)
	}
}
