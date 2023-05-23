package model

import (
	geo "github.com/kellydunn/golang-geo"
	"github.com/samber/lo"
	"math/rand"
)

type Parking struct {
	Name   string         `json:"name"`
	Lon    float64        `json:"lon"`
	Lat    float64        `json:"lat"`
	Radius float64        `json:"radius"`
	Trucks map[int]*Truck `json:"trucks"`
}

func (p *Parking) AddTruck(truck *Truck, space *StartWorkSpace) {
	truck.Lat = p.Lat
	truck.Lat = p.Lon
	truck.Status = 0
	truck.Route = truck.DpsSimulator(geo.NewPoint(space.Lat, space.Lon), geo.NewPoint(space.Lat, space.Lon))
	p.Trucks[truck.Id] = truck
}

func (p *Parking) RandReduceTruck(num int) map[int]*Truck {
	machineIds := lo.Keys(p.Trucks)
	res := make(map[int]*Truck)
	for i := 0; i < num; i++ {
		key := rand.Intn(len(machineIds)-1) + 1
		res[key] = p.Trucks[machineIds[key]]
		p.ReduceTruck(res[key])
		machineIds = append(machineIds[:key], machineIds[key+1:]...)
	}

	return res
}

func (p *Parking) ReduceTruck(truck *Truck) {
	delete(p.Trucks, truck.Id)
}
