package service

type CreateIncedentRequestInput struct {
	Title       string
	Description *string
	Lat         float64
	Long        float64
	Radius      int
	Active      *bool
}

type IncedentRequestDTO struct {
	ID     int
	Title  string
	Lat    float64
	Long   float64
	Radius int
	Active bool
}

type CreateIncedentOutput struct {
	Incedent IncedentRequestDTO
}
