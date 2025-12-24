package handler

type incedentRequestJSON struct {
	ID     int     `db:"id" json:"id"`
	Title  string  `db:"title" json:"title"`
	Lat    float64 `db:"lat" json:"lat"`
	Long   float64 `db:"long" json:"long"`
	Radius int     `db:"radius_m" json:"redius_m"`
	Active bool    `db:"active" json:"active"`
}

type newIncedentJSON struct {
	Title       string  `json:"title"`
	Description *string `json:"description,omitempty"`
	Lat         float64 `json:"lat"`
	Long        float64 `json:"long"`
	Radius      int     `json:"radius_m"`
	Active      *bool   `json:"active,omitempty"`
}

type getIncedentJSON struct {
	ID int `json:"id"`
}

type changeIncedentJSON struct {
	Lat    float64 `json:"lat"`
	Long   float64 `json:"long"`
	Radius *int    `json:"radius_m"`
}

// Responses

type incedentRequestResponse struct {
	Incendent incedentRequestJSON `json:"Incedent"`
}
