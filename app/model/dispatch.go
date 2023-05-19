package model

type Dispatch struct {
	Parking   Parking    `json:"parking"`
	Platforms []Platform `json:"platforms"`
}
