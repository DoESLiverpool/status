package database

import "time"

// ServiceState is the possible state a service can be in
type ServiceState int

const (
	// UnkownState is for when we can not determin if the service is working
	UnkownState ServiceState = 0

	// BrokenState is for when we know a service is broken
	BrokenState ServiceState = 1

	// WorkingState is for when we know a service is working
	WorkingState ServiceState = 2
)

// Service is a service stored in the database
type Service struct {
	ID          int64        `json:"id" binding:"required"`
	Name        string       `json:"name" binding:"required"`
	State       ServiceState `json:"state"`
	Description string       `json:"description"`
	Since       time.Time    `json:"since"`
}

// ServiceBucket is the name of the bucket that contains all the services
const ServiceBucket = "services"
