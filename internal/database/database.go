package database

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"vn-admin-api/internal/config"
	"vn-admin-api/internal/models"

	_ "github.com/lib/pq"
)

// Repository handles all database interactions
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new Repository instance
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Connect establishes a connection to the database and returns a Repository
func Connect(cfg *config.Config) (*Repository, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode,
	)

	// Open connection (lazy)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Verify connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close() // cleanup
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Production: Configure connection pool for high load
	db.SetMaxOpenConns(100)                // Max concurrent connections (for 1000 req/s)
	db.SetMaxIdleConns(50)                 // Keep more idle connections ready
	db.SetConnMaxLifetime(5 * time.Minute) // Recycle connections

	return &Repository{db: db}, nil
}

// Close closes the database connection
func (r *Repository) Close() error {
	return r.db.Close()
}

// DB returns the underlying sql.DB (useful for transactions or direct access if needed)
func (r *Repository) DB() *sql.DB {
	return r.db
}

// InitSchema executes the schema SQL to create tables
func (r *Repository) InitSchema(schemaSQL string) error {
	if _, err := r.db.Exec(schemaSQL); err != nil {
		return fmt.Errorf("failed to execute schema: %w", err)
	}
	return nil
}

// UpsertProvince inserts or updates a province
func (r *Repository) UpsertProvince(ctx context.Context, p models.Province) error {
	query := `
		INSERT INTO provinces (id, name, code, updated_at) 
		VALUES ($1, $2, $3, NOW()) 
		ON CONFLICT (id) 
		DO UPDATE SET name = $2, code = $3, updated_at = NOW()`

	_, err := r.db.ExecContext(ctx, query, p.ID, p.Name, fmt.Sprintf("%d", p.Code))
	if err != nil {
		return fmt.Errorf("failed to upsert province %d: %w", p.ID, err)
	}
	return nil
}

// UpsertAdminUnit inserts or updates an admin unit
func (r *Repository) UpsertAdminUnit(ctx context.Context, u models.AdminUnit) error {
	query := `
		INSERT INTO admin_units (id, province_id, name, level, code, pre_merger_desc, lat, long, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW()) 
		ON CONFLICT (id) 
		DO UPDATE SET 
			name=$3, level=$4, code=$5, pre_merger_desc=$6, lat=$7, long=$8, updated_at=NOW()`

	_, err := r.db.ExecContext(ctx, query,
		u.ID, u.ProvinceID, u.Name, u.Level, u.Code, u.PreMergerDesc, u.Lat, u.Long)
	if err != nil {
		return fmt.Errorf("failed to upsert unit %d: %w", u.ID, err)
	}
	return nil
}

// GetProvinces returns all provinces
func (r *Repository) GetProvinces(ctx context.Context) ([]models.Province, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, name, code, updated_at FROM provinces ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	provinces := make([]models.Province, 0)
	for rows.Next() {
		var p models.Province
		var codeStr sql.NullString

		if err := rows.Scan(&p.ID, &p.Name, &codeStr, &p.UpdatedAt); err != nil {
			return nil, err
		}
		if codeStr.Valid {
			if codeVal, err := strconv.Atoi(codeStr.String); err == nil {
				p.Code = codeVal
			}
		}
		provinces = append(provinces, p)
	}
	return provinces, nil
}

// GetUnitsByProvince returns all admin units for a specific province
func (r *Repository) GetUnitsByProvince(ctx context.Context, provinceID int) ([]models.AdminUnit, error) {
	rows, err := r.db.QueryContext(ctx,
		"SELECT id, province_id, name, level, code, pre_merger_desc, lat, long, updated_at FROM admin_units WHERE province_id = $1 ORDER BY id",
		provinceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	units := make([]models.AdminUnit, 0)
	for rows.Next() {
		var u models.AdminUnit
		var preMerger sql.NullString

		if err := rows.Scan(&u.ID, &u.ProvinceID, &u.Name, &u.Level, &u.Code, &preMerger, &u.Lat, &u.Long, &u.UpdatedAt); err != nil {
			return nil, err
		}
		if preMerger.Valid {
			u.PreMergerDesc = preMerger.String
		}
		units = append(units, u)
	}
	return units, nil
}

// SearchAdminUnits searches for units by name or pre-merger description
func (r *Repository) SearchAdminUnits(ctx context.Context, query string) ([]models.AdminUnit, error) {
	// Simple ILIKE search. For better performance with large data, consider Full Text Search (tsvector).
	sqlQuery := `
		SELECT id, province_id, name, level, code, pre_merger_desc, lat, long, updated_at 
		FROM admin_units 
		WHERE name ILIKE '%' || $1 || '%' OR pre_merger_desc ILIKE '%' || $1 || '%' 
		ORDER BY id 
		LIMIT 50`

	rows, err := r.db.QueryContext(ctx, sqlQuery, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	units := make([]models.AdminUnit, 0)
	for rows.Next() {
		var u models.AdminUnit
		var preMerger sql.NullString

		if err := rows.Scan(&u.ID, &u.ProvinceID, &u.Name, &u.Level, &u.Code, &preMerger, &u.Lat, &u.Long, &u.UpdatedAt); err != nil {
			return nil, err
		}
		if preMerger.Valid {
			u.PreMergerDesc = preMerger.String
		}
		units = append(units, u)
	}
	return units, nil
}
