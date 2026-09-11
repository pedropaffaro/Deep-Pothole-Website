package complaint

import "io"

type Complaint struct {
	ID        uint
	City      string
	Street    string
	Latitude  float64
	Longitude float64
	PhotoKey  string
}

type CreateInput struct {
	City        string
	Street      string
	Latitude    float64
	Longitude   float64
	Photo       io.Reader
	PhotoName   string
	ContentType string
}
