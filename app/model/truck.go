package model

import (
	geo "github.com/kellydunn/golang-geo"
	"math"
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

func (t *Truck) Load(space *StartWorkSpace) {
	t.Route = t.DpsSimulator(geo.NewPoint(t.Lat, t.Lon), geo.NewPoint(space.Lat, space.Lon))
	t.Lat = space.Lat
	t.Lon = space.Lon
	t.Status = 1

}

func (t *Truck) Unload(space *EndWorkSpace) {
	t.Route = t.DpsSimulator(geo.NewPoint(t.Lat, t.Lon), geo.NewPoint(space.Lat, space.Lon))
	t.Lat = space.Lat
	t.Lon = space.Lon
	t.Status = 3

}

func (t *Truck) DpsSimulator(start *geo.Point, end *geo.Point) (res []Route) {
	// 两个GPS点之间的距离
	bearing := start.BearingTo(end) // 两个GPS点之间的方向角
	// 每500米生成一个GPS点
	distance := start.GreatCircleDistance(end) / 10 // 两个GPS点之间的距离
	res = append(res, Route{Lat: start.Lat(), Lon: start.Lng()})
	for i := 1; i < 10; i++ {
		// 计算每个中间点的经纬度
		mp := t.pointOnBearing(start, bearing, float64(i)*distance)
		res = append(res, Route{Lat: mp.Lat(), Lon: mp.Lng()})
	}
	res = append(res, Route{Lat: end.Lat(), Lon: end.Lng()})

	return res
}

func (t *Truck) pointOnBearing(p *geo.Point, bearing, distance float64) *geo.Point {
	// 计算新GPS点的经度和纬度
	lat1 := p.Lat() * math.Pi / 180.0
	lon1 := p.Lng() * math.Pi / 180.0
	b := bearing * math.Pi / 180.0
	d := distance / 6378.1 // 地球半径约为6378.1千米
	lat2 := math.Asin(math.Sin(lat1)*math.Cos(d) + math.Cos(lat1)*math.Sin(d)*math.Cos(b))
	lon2 := lon1 + math.Atan2(math.Sin(b)*math.Sin(d)*math.Cos(lat1), math.Cos(d)-math.Sin(lat1)*math.Sin(lat2))
	// 将经度和纬度转换回弧度制
	lat2 = lat2 * 180.0 / math.Pi
	lon2 = lon2 * 180.0 / math.Pi
	return geo.NewPoint(lat2, lon2)
}
