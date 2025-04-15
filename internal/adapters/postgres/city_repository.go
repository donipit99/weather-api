package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"weather-api/internal/models"
)

type CityRepository struct {
	db *sql.DB
}

func NewCityRepository(db *sql.DB) *CityRepository {
	return &CityRepository{db: db}
}

func (r *CityRepository) GetCityByName(ctx context.Context, name string) (*models.City, error) {
	query := `SELECT name, latitude, longitude, country FROM cities WHERE name = $1`
	var city models.City
	err := r.db.QueryRowContext(ctx, query, name).Scan(&city.Name, &city.Latitude, &city.Longitude, &city.Country)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("city not found: %s", name)
	}
	if err != nil {
		slog.Error("failed to query city", "name", name, "err", err)
		return nil, fmt.Errorf("db.QueryRowContext: %w", err)
	}
	return &city, nil
}

func (r *CityRepository) GetAllCities(ctx context.Context) ([]models.City, error) {
	query := `SELECT name, latitude, longitude, country FROM cities`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		slog.Error("failed to query cities", "err", err)
		return nil, fmt.Errorf("db.QueryContext: %w", err)
	}
	defer rows.Close()

	var cities []models.City
	for rows.Next() {
		var city models.City
		if err := rows.Scan(&city.Name, &city.Latitude, &city.Longitude, &city.Country); err != nil {
			slog.Error("failed to scan city", "err", err)
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}
		cities = append(cities, city)
	}
	return cities, nil
}
