package models

type Complaint struct{
	ID			uint	`json:"id"`
	City 		string 	`json:"city"`
	Street 		string 	`json:"street"`
	Latitude 	float64 `json:"latitude"`
	Longitude 	float64 `json:"longitude"`
	PhotoURL 	string 	`json:"photo_url"`
}