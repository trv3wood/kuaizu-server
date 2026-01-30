package repository

import (
	"github.com/jmoiron/sqlx"
)

// Repository aggregates all sub-repositories
type Repository struct {
	User        *UserRepository
	Project     *ProjectRepository
	Product     *ProductRepository
	Application *ApplicationRepository
	OliveBranch *OliveBranchRepository
	School      *SchoolRepository
	Major       *MajorRepository
}

// New creates a new Repository with all sub-repositories
func New(db *sqlx.DB) *Repository {
	return &Repository{
		User:        NewUserRepository(db),
		Project:     NewProjectRepository(db),
		Product:     NewProductRepository(db),
		Application: NewApplicationRepository(db),
		OliveBranch: NewOliveBranchRepository(db),
		School:      NewSchoolRepository(db),
		Major:       NewMajorRepository(db),
	}
}
