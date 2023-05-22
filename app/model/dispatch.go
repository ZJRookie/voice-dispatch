package model

type Dispatch struct {
	Parking   *Parking          `json:"parking"`
	Platforms map[int]*Platform `json:"platforms"`
}
