package model

type Platform struct {
	Id             int            `json:"id"`
	StartWorkSpace StartWorkSpace `json:"start_work_space"`
	EndWorkSpace   EndWorkSpace   `json:"end_work_space"`
}
