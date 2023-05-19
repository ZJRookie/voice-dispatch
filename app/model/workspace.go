package model

type WorkSpace struct {
	Id     int     `json:"id"`
	Name   string  `json:"name"`
	Lon    float64 `json:"lon"`
	Lat    float64 `json:"lat"`
	Radius float64 `json:"radius"`
}

type StartWorkSpace struct {
	WorkSpace
	Excavator Excavator `json:"excavator"`
}

type EndWorkSpace struct {
	WorkSpace
}
