package database

import (
	"encoding/json"
	"errors"
	"time"

	bolt "github.com/coreos/bbolt"
)

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

// ServiceHelper is a helper for all service methods
type ServiceHelper struct {
	store *Store
}

// CreateService will create a new or update a service in the database
func (s *ServiceHelper) CreateService(u *Service) error {
	return s.store.db.Update(func(tx *bolt.Tx) error {
		// Retrieve the services bucket.
		// This should be created when the DB is first opened.
		b := tx.Bucket([]byte(ServiceBucket))

		// Marshal user data into bytes.
		buf, err := json.Marshal(*u)
		if err != nil {
			return err
		}

		// Persist bytes to users bucket.
		return b.Put(idToKey(u.ID), buf)
	})
}

// GetServices will return all the services in the database
func (s *ServiceHelper) GetServices() ([]*Service, error) {
	var services []*Service
	err := s.store.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(ServiceBucket))

		return b.ForEach(func(k, d []byte) error {

			var service = Service{}

			err := json.Unmarshal(d, &service)

			if err != nil {
				return err
			}

			services = append(services, &service)

			return nil
		})
	})

	if err != nil {
		return nil, err
	}

	return services, nil
}

// GetService will return the service by id
func (s *ServiceHelper) GetService(id int64) (*Service, error) {
	var service = Service{}
	err := s.store.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(ServiceBucket))

		d := b.Get(idToKey(id))

		if d == nil {
			return errors.New("No service found")
		}

		err := json.Unmarshal(d, &service)

		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &service, nil
}

// UpdateServices will update all the services in the db and add new ones
func (s *ServiceHelper) UpdateServices(services []*Service) error {
	for _, service := range services {
		serv, err := s.GetService(service.ID)

		// If the service isn't new do state checks
		if serv != nil {

			// If state matches set old time.
			// Else create a history event for the service changing
			if serv.State == service.State {
				service.Since = serv.Since
			} else {
				var history = &History{
					OldTimestamp: serv.Since,
					NewTimestamp: service.Since,

					ChangedFrom: serv.State,
					ChangedTo:   serv.State,

					ServiceID: service.ID,
				}

				s.store.History.CreateHistoryEvent(history)
			}
		}

		err = s.CreateService(service)

		if err != nil {
			return err
		}
	}

	return nil
}
