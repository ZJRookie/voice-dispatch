package request

type NotifyRequest struct {
	Machines []DispatchMachine `json:"machines"`
}

type DispatchMachine struct {
	Sn      string `json:"sn"`
	Message string `json:"message"`
}
