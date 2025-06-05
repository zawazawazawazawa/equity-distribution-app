package main

import (
	"equity-distribution-backend/pkg/utils"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	// データベース接続情報を環境変数から取得
	host := utils.GetEnvOrDefault("POSTGRES_HOST", "localhost")
	port := utils.GetEnvIntOrDefault("POSTGRES_PORT", 5432)
	user := utils.GetEnvOrDefault("POSTGRES_USER", "postgres")
	password := utils.GetEnvOrDefault("POSTGRES_PASSWORD", "postgres")
	dbName := utils.GetEnvOrDefault("POSTGRES_DBNAME", "plo_equity")

	// データベース接続文字列を構築
	databaseURL := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		user, password, host, port, dbName)

	// マイグレーションインスタンスを作成
	m, err := migrate.New(
		"file://migrations",
		databaseURL,
	)
	if err != nil {
		log.Fatalf("Failed to create migrate instance: %v", err)
	}
	defer m.Close()

	switch command {
	case "up":
		err = m.Up()
		if err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Failed to run migrations: %v", err)
		}
		if err == migrate.ErrNoChange {
			fmt.Println("No migrations to apply")
		} else {
			fmt.Println("Migrations applied successfully")
		}

	case "down":
		err = m.Steps(-1)
		if err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Failed to rollback migration: %v", err)
		}
		if err == migrate.ErrNoChange {
			fmt.Println("No migrations to rollback")
		} else {
			fmt.Println("Migration rolled back successfully")
		}

	case "status":
		version, dirty, err := m.Version()
		if err != nil {
			if err == migrate.ErrNilVersion {
				fmt.Println("No migrations have been applied")
			} else {
				log.Fatalf("Failed to get migration status: %v", err)
			}
		} else {
			fmt.Printf("Current migration version: %d\n", version)
			if dirty {
				fmt.Println("WARNING: Database is in dirty state")
			} else {
				fmt.Println("Database is clean")
			}
		}

	case "version":
		if len(os.Args) < 3 {
			fmt.Println("Please specify target version")
			printUsage()
			os.Exit(1)
		}
		targetVersion, err := strconv.ParseUint(os.Args[2], 10, 32)
		if err != nil {
			log.Fatalf("Invalid version number: %v", err)
		}
		err = m.Migrate(uint(targetVersion))
		if err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Failed to migrate to version %d: %v", targetVersion, err)
		}
		if err == migrate.ErrNoChange {
			fmt.Printf("Already at version %d\n", targetVersion)
		} else {
			fmt.Printf("Migrated to version %d successfully\n", targetVersion)
		}

	case "force":
		if len(os.Args) < 3 {
			fmt.Println("Please specify version to force")
			printUsage()
			os.Exit(1)
		}
		version, err := strconv.ParseInt(os.Args[2], 10, 32)
		if err != nil {
			log.Fatalf("Invalid version number: %v", err)
		}
		err = m.Force(int(version))
		if err != nil {
			log.Fatalf("Failed to force version %d: %v", version, err)
		}
		fmt.Printf("Forced database to version %d\n", version)

	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: go run cmd/migrate/main.go <command> [args]")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  up              Apply all pending migrations")
	fmt.Println("  down            Rollback the last migration")
	fmt.Println("  status          Show current migration status")
	fmt.Println("  version <n>     Migrate to specific version")
	fmt.Println("  force <n>       Force database to specific version (use with caution)")
	fmt.Println("")
	fmt.Println("Environment variables:")
	fmt.Println("  POSTGRES_HOST     (default: localhost)")
	fmt.Println("  POSTGRES_PORT     (default: 5432)")
	fmt.Println("  POSTGRES_USER     (default: postgres)")
	fmt.Println("  POSTGRES_PASSWORD (default: postgres)")
	fmt.Println("  POSTGRES_DBNAME   (default: plo_equity)")
}

