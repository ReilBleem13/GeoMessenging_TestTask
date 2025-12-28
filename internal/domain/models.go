package domain

import (
	"time"
)

type Incident struct {
	ID          int       `db:"id" json:"id"`
	Title       string    `db:"title" json:"title"`
	Description string    `db:"description" json:"description"`
	Lat         float64   `db:"lat" json:"lat"`
	Long        float64   `db:"long" json:"long"`
	Radius      int       `db:"radius_m" json:"radius_m"`
	Active      bool      `db:"active" json:"active"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

type LocationCheck struct {
	ID           int       `db:"id" json:"id"`
	UserID       string    `db:"user_id" json:"user_id"`
	CheckedAt    time.Time `db:"checked_at" json:"checked_at"`
	Lat          float64   `db:"lat" json:"lat"`
	Long         float64   `db:"long" json:"long"`
	InDangerZone bool      `db:"in_danger_zone" json:"in_danger_zone"`
	NearestID    *int      `db:"nearest_id" json:"nearest_id,omitempty"`
}

type ZoneStat struct {
	ZoneID    int `db:"zone_id" json:"zone_id"`
	UserCount int `db:"user_count" json:"user_count"`
}
