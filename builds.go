package conveyor

import (
	"fmt"

	"code.google.com/p/go-uuid/uuid"
	"github.com/jinzhu/gorm"
)

type Build struct {
	// Unique identifier for this build.
	ID string

	// Identifier is the unique identifier for the build image.
	Identifier string

	// The options provided to start the build.
	BuildOptions
}

// Remote is the value that you would use when pulling this image.
func (b *Build) Remote() string {
	return fmt.Sprintf("%s:%s", b.Repository, b.Identifier)
}

func (b *Build) BeforeCreate() error {
	b.ID = uuid.New()
	return nil
}

// BuildsService is a gorm.DB backed persistence layer for Builds.
type BuildsService struct {
	db *gorm.DB
}

// Create persists the build.
func (s *BuildsService) Create(b *Build) error {
	return s.db.Create(b).Error
}
