package models

import "time"

// Area represents a geographic grouping for climbing locations
type Area struct {
	ID           int       `json:"id" db:"id"`
	Name         string    `json:"name" db:"name"`
	Description  *string   `json:"description,omitempty" db:"description"`
	Region       *string   `json:"region,omitempty" db:"region"`
	DisplayOrder int       `json:"display_order" db:"display_order"`
	IsActive     bool      `json:"is_active" db:"is_active"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// AreaWithLocationCount extends Area with location count
type AreaWithLocationCount struct {
	Area
	LocationCount int `json:"location_count" db:"location_count"`
}
