package model

type Parking struct {
	Name   string  `json:"name"`
	Lon    float64 `json:"lon"`
	Lat    float64 `json:"lat"`
	Radius float64 `json:"radius"`
	Trucks []Truck `json:"trucks"`
}
