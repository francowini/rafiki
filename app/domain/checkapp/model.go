package checkapp

import "encoding/json"

// Info represents information about the service.
type Info struct {
	Host string `json:"host,omitempty"`
	Status     string `json:"status,omitempty"`
	Build      string `json:"build,omitempty"`
	GOMAXPROCS int    `json:"GOMAXPROCS,omitempty"`
}

// Encode implements the encoder interface.
func (app Info) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}
