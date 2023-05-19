package model

type Truck struct {
	Id   int     `json:"id"`
	Name string  `json:"name"`
	Sn   string  `json:"sn"`
	Lon  float64 `json:"lon"`
	Lat  float64 `json:"lat"`
}