package services

import (
	database "github.com/DoESLiverpool/status/database"
)

// Gets all of the doorbot services. Marks broken ones that haven't pinged since the timeout
func getDoorbotServices() *[]database.Service {
	return nil
}
