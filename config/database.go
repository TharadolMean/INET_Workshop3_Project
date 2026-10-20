package config

import (
	"database/sql"
	"fmt"

	"employee-profile-api/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	_ "modernc.org/sqlite"
)

func OpenDatabase(path string) (*gorm.DB, error) {
	// modernc's sqlite driver is pure Go; the GORM dialector uses its registered name.
	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open SQLite: %w", err)
	}
	conn.SetMaxOpenConns(1)
	db, err := gorm.Open(sqlite.New(sqlite.Config{Conn: conn, DriverName: "sqlite"}), &gorm.Config{TranslateError: true})
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("connect SQLite: %w", err)
	}
	if err := db.AutoMigrate(&models.Employee{}); err != nil {
		conn.Close()
		return nil, fmt.Errorf("migrate employees: %w", err)
	}
	return db, nil
}
