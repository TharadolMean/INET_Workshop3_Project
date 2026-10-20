package main

import (
	"log"
	"os"

	"employee-profile-api/config"
	"employee-profile-api/routes"
	"employee-profile-api/seed"
)

func main() {
	dbPath := envOr("DB_PATH", "employee.db")
	db, err := config.OpenDatabase(dbPath)
	if err != nil {
		log.Fatal(err)
	}
	if err := seed.Employees(db); err != nil {
		log.Fatal(err)
	}
	app := routes.NewApp(db, envOr("BASIC_AUTH_USER", "testgo"), envOr("BASIC_AUTH_PASSWORD", "23012023"))
	log.Fatal(app.Listen(":3000"))
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
