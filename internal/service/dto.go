package service

type CreateIncedentRequestInput struct {
	Title       string
	Description *string
	Lat         float64
	Long        float64
	Radius      int
	Active      *bool
}
