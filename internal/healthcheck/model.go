package healthcheck

const ServiceName = "micro-stakes-api"

type Health struct {
	ServiceName string `json:"service_name"`
	Status      string `json:"status"`
	Message     string `json:"message"`
}
