package model

import (
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

func (p *Parking) RandReduceTruck(num int) (res map[int]*Truck) {
	machineIds := lo.Keys(p.Trucks)
	for i := 0; i < num; i++ {
		key := rand.Intn(len(machineIds))
		res[key] = p.Trucks[machineIds[key]]
		p.ReduceTruck(p.Trucks[key])
		machineIds = append(machineIds[:key], machineIds[key+1:]...)
	}

	return
}

func (p *Parking) ReduceTruck(truck *Truck) {
	delete(p.Trucks, truck.Id)
}
