package store

import (
	"fmt"
	"os"
	"path/filepath"
	"github.com/kjbreil/go-plex/library"
	"github.com/kjbreil/mediate/pkg/shows"
	_ "github.com/mattn/go-sqlite3"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Store struct {
	gorm.DB
	libraries map[string]*library.Library
}

func InitDB() (*Store, error) {
	return InitDBWithPath("mediate.sqlite")
}

func InitDBWithPath(dbPath string) (*Store, error) {
	s := &Store{
		libraries: make(map[string]*library.Library),
	}
	
	// Convert to absolute path if it's relative
	if !filepath.IsAbs(dbPath) {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %v", err)
		}
		dbPath = filepath.Join(homeDir, ".local", "share", "mediate", dbPath)
	}
	
	// Create directory if it doesn't exist
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %v", err)
	}
	
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	s.DB = *db

	err = s.DB.AutoMigrate(
		&shows.Show{},
		&shows.Episode{},
		&shows.ViewingSession{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to run database migrations: %v", err)
	}
	
	return s, nil
}

func (s *Store) AddLibrary(lib *library.Library) {
	s.libraries[lib.UUID] = lib
}
