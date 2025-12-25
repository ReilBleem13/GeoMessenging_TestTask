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
