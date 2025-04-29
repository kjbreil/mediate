package store

import (
	"fmt"
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

	s := &Store{
		libraries: make(map[string]*library.Library),
	}
	var err error
	db, err := gorm.Open(sqlite.Open("mediate.sqlite"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	s.DB = *db

	err = s.DB.AutoMigrate(
		&shows.Show{},
		&shows.Episode{},
	)
	return s, nil
}

func (s *Store) AddLibrary(lib *library.Library) {
	s.libraries[lib.UUID] = lib
}
