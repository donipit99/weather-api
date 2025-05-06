package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"weather-api/internal/models"
	"weather-api/internal/repository"

	"github.com/jmoiron/sqlx"
)

// Убедимся, что CityRepository реализует интерфейс repository.CityRepository
var _ repository.CityRepository = (*CityRepository)(nil)

type CityRepository struct {
	db *sqlx.DB
}

type CityRepositoryOptions struct {
	DB *sqlx.DB
}

func NewCityRepository(options CityRepositoryOptions) *CityRepository {
	return &CityRepository{db: options.DB}
}

// GetCityByName получает город по имени из PostgreSQL
func (r *CityRepository) GetCityByName(ctx context.Context, name string) (*models.City, error) {
	// Реализация запроса к PostgreSQL
	query := `SELECT name, latitude, longitude, country FROM cities WHERE name = $1`

	var city models.City
	err := r.db.QueryRowContext(ctx, query, name).Scan(
		&city.Name,
		&city.Latitude,
		&city.Longitude,
		&city.Country,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("city not found: %s", name)
		}
		return nil, fmt.Errorf("query error: %w", err)
	}

	return &city, nil
}

// GetAllCities получает все города из PostgreSQL
func (r *CityRepository) GetAllCities(ctx context.Context) ([]models.City, error) {
	// Реализация запроса к PostgreSQL
	query := `SELECT name, latitude, longitude, country FROM cities`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}
	defer rows.Close()

	var cities []models.City
	for rows.Next() {
		var city models.City
		if err := rows.Scan(
			&city.Name,
			&city.Latitude,
			&city.Longitude,
			&city.Country,
		); err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}
		cities = append(cities, city)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return cities, nil
}
