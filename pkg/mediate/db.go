package mediate

import (
	"fmt"
	"github.com/kjbreil/mediate/pkg/shows"
	_ "github.com/mattn/go-sqlite3"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func (m *Mediate) initDB() error {
	db, err := gorm.Open(sqlite.Open("mediate.sqlite"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})

	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	err = db.AutoMigrate(
		&shows.DbShow{},
		&shows.DbEpisode{},
	)
	return nil
}

// func (m *Mediate) initDB() error {
//	var err error
//	m.db, err = sql.Open("sqlite3", "mediate.sqlite")
//
//	if err != nil {
//		fmt.Errorf("failed to connect to database: %v", err)
//	}
//
//	m.queries = model.New(m.db)
//
//	return nil
// }
