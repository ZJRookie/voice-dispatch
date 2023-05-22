package model

import (
	geo "github.com/kellydunn/golang-geo"
	"voice-dispatch/app/helper"
)

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

func (t *Truck) Load(space *EndWorkSpace) {
	t.Lat = space.Lat
	t.Lon = space.Lon
}

func (t *Truck) Unload(lat, lon float64, space *StartWorkSpace) {
	t.Lat = space.Lat
	t.Lon = space.Lon
	t.Status = 1
	t.Route = helper.DpsSimulator(geo.NewPoint(lat, lon), geo.NewPoint(space.Lat, space.Lon))
}
