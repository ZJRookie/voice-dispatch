package model

type Truck struct {
	Id     int     `json:"id"`
	Name   string  `json:"name"`
	Sn     string  `json:"sn"`
	Lon    float64 `json:"lon"`
	Lat    float64 `json:"lat"`
	Route  []Route `json:"route"`
	Status int     `json:"status"` // 0 停车场  1 装载 2 运输中 3 卸载
}

type Route struct {
	Lon float64 `json:"lon"`
	Lat float64 `json:"lat"`
}
