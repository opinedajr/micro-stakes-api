package healthcheck

func CheckHealth() []Health {
	return []Health{
		{
			ServiceName: ServiceName,
			Status:      "healthy",
			Message:     "Service is running",
		},
	}
}
