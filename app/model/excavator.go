package model

type Excavator struct {
	Id              int     `json:"id"`
	Name            string  `json:"name"`
	Lon             float64 `json:"lon"`
	Lat             float64 `json:"lat"`
	Status          string  `json:"status"`
	Trucks          []Truck `json:"trucks"`
	CurrentTruckNum int     `json:"current_truck_num"`
	MaxTruckNum     int     `json:"max_truck_num,"`
}
