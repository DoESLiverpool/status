package services

import "github.com/DoESLiverpool/status/database"

// Find each service by their ID
func getServiceByID(services []*database.Service, id int64) *database.Service {
	for _, service := range services {
		if service.ID == id {
			return service
		}
	}

	return nil
}
