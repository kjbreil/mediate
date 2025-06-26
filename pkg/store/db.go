package store

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
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
		&shows.DeletedMedia{},
		&shows.DeletedMediaSession{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to run database migrations: %v", err)
	}
	
	return s, nil
}

func (s *Store) AddLibrary(lib *library.Library) {
	s.libraries[lib.UUID] = lib
}

// SaveDeletedMedia saves a deleted media record to the database
func (s *Store) SaveDeletedMedia(media *shows.DeletedMedia) error {
	// Check if media already exists
	var existing shows.DeletedMedia
	result := s.DB.Where("rating_key = ?", media.RatingKey).First(&existing)
	
	if result.Error == nil {
		// Update existing record
		media.ID = existing.ID
		media.CreatedAt = existing.CreatedAt
		return s.DB.Save(media).Error
	}
	
	// Create new record
	return s.DB.Create(media).Error
}

// GetDeletedMedia retrieves deleted media records with optional filters
func (s *Store) GetDeletedMedia(limit int, offset int) ([]*shows.DeletedMedia, error) {
	var deletedMedia []*shows.DeletedMedia
	query := s.DB.Preload("ViewingSessions").Order("deleted_at DESC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	
	err := query.Find(&deletedMedia).Error
	return deletedMedia, err
}

// GetDeletedMediaByRatingKey retrieves a specific deleted media record
func (s *Store) GetDeletedMediaByRatingKey(ratingKey string) (*shows.DeletedMedia, error) {
	var media shows.DeletedMedia
	err := s.DB.Preload("ViewingSessions").Where("rating_key = ?", ratingKey).First(&media).Error
	if err != nil {
		return nil, err
	}
	return &media, nil
}

// GetDeletedMediaSummary generates summary statistics for deleted media
func (s *Store) GetDeletedMediaSummary() (*shows.DeletedMediaSummary, error) {
	var totalItems int64
	var totalViews int64
	
	// Count total items
	if err := s.DB.Model(&shows.DeletedMedia{}).Count(&totalItems).Error; err != nil {
		return nil, err
	}
	
	// Sum total views
	if err := s.DB.Model(&shows.DeletedMedia{}).Select("COALESCE(SUM(total_views), 0)").Scan(&totalViews).Error; err != nil {
		return nil, err
	}
	
	// Count items deleted in last 30 days
	var recentDeletions int64
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	if err := s.DB.Model(&shows.DeletedMedia{}).Where("deleted_at > ?", thirtyDaysAgo).Count(&recentDeletions).Error; err != nil {
		return nil, err
	}
	
	// Group by media type
	var mediaTypeCounts []struct {
		MediaType string
		Count     int
	}
	if err := s.DB.Model(&shows.DeletedMedia{}).Select("media_type, COUNT(*) as count").Group("media_type").Scan(&mediaTypeCounts).Error; err != nil {
		return nil, err
	}
	
	byMediaType := make(map[string]int)
	for _, item := range mediaTypeCounts {
		byMediaType[item.MediaType] = item.Count
	}
	
	// Group by library
	var libraryCounts []struct {
		LibrarySectionID int
		Count            int
	}
	if err := s.DB.Model(&shows.DeletedMedia{}).Select("library_section_id, COUNT(*) as count").Group("library_section_id").Scan(&libraryCounts).Error; err != nil {
		return nil, err
	}
	
	byLibrary := make(map[int]int)
	for _, item := range libraryCounts {
		byLibrary[item.LibrarySectionID] = item.Count
	}
	
	// Get top deleted shows by view count
	var topDeleted []*shows.DeletedMedia
	if err := s.DB.Order("total_views DESC").Limit(10).Find(&topDeleted).Error; err != nil {
		return nil, err
	}
	
	var topDeletedShows []shows.DeletedMediaStats
	for _, media := range topDeleted {
		topDeletedShows = append(topDeletedShows, *media.GetViewingStats())
	}
	
	// Get recent deletions
	var recentDeletedMedia []*shows.DeletedMedia
	if err := s.DB.Order("deleted_at DESC").Limit(20).Find(&recentDeletedMedia).Error; err != nil {
		return nil, err
	}
	
	return &shows.DeletedMediaSummary{
		TotalItems:      int(totalItems),
		TotalViews:      int(totalViews),
		DeletedRecently: int(recentDeletions),
		ByMediaType:     byMediaType,
		ByLibrary:       byLibrary,
		TopDeletedShows: topDeletedShows,
		RecentDeletions: recentDeletedMedia,
		GeneratedAt:     time.Now(),
	}, nil
}
