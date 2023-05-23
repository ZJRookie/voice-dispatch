package model

type Excavator struct {
	Id              int            `json:"id"`
	Name            string         `json:"name"`
	Lon             float64        `json:"lon"`
	Lat             float64        `json:"lat"`
	Status          string         `json:"status"`
	Trucks          map[int]*Truck `json:"trucks"`
	CurrentTruckNum int            `json:"current_truck_num"`
	MaxTruckNum     int            `json:"max_truck_num,"`
}

func (e *Excavator) AddTrucks(platform *Platform, trucks map[int]*Truck) {
	for _, truck := range trucks {
		e.AddTruck(platform, truck)
	}
}
func (e *Excavator) AddTruck(platform *Platform, truck *Truck) {
	e.Trucks[truck.Id] = truck
	e.CurrentTruckNum++
	truck.Load(platform.StartWorkSpace)
}

func (e *Excavator) ReduceTruck(truck *Truck) {
	delete(e.Trucks, truck.Id)
	e.CurrentTruckNum--
}
