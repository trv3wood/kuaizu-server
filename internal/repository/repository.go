package repository

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository aggregates all sub-repositories
type Repository struct {
	User        *UserRepository
	Project     *ProjectRepository
	Product     *ProductRepository
	Application *ApplicationRepository
}

// New creates a new Repository with all sub-repositories
func New(pool *pgxpool.Pool) *Repository {
	return &Repository{
		User:        NewUserRepository(pool),
		Project:     NewProjectRepository(pool),
		Product:     NewProductRepository(pool),
		Application: NewApplicationRepository(pool),
	}
}
