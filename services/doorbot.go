package services

import (
	"strconv"
	"time"

	"github.com/caarlos0/env"

	database "github.com/DoESLiverpool/status/database"
)

// DoorbotSettings is the model for all the environment settings used for the doorbot service
type DoorbotSettings struct {
	Disabled string `env:"DOORBOT_DISABLED"`
	APIKey   string `env:"DOORBOT_API_KEY"`
	//Timeout string `env:"DOORBOT_TIMEOUT"`
}

// Doorbot is the model that is used when the api recieves an update
type Doorbot struct {
	ID int64 `json:"id" binding:"required"`

	Name      string    `json:"name"`
	Timestamp time.Time `json:"timestamp" binding:"required"`
}

var updatedDoorbots []*Doorbot
var notFoundDoorbots []*Doorbot

// UpdateDoorbots will return all the doorbots in the state that the current system sees them in
func UpdateDoorbots() ([]*database.Service, error) {
	var settings = DoorbotSettings{}
	env.Parse(&settings)

	if settings.Disabled != "" && settings.Disabled != "false" {
		return nil, nil
	}

	store := database.Store{}
	err := store.GetDatabase(true)
	defer store.CloseDatabase()

	if err != nil {
		return nil, err
	}

	doorbots, err := getExistingDoorbots(&store)

	if err != nil {
		return nil, err
	}

	var foundDoorbots []*database.Service

	for _, service := range doorbots {
		var found = false
		for i, door := range updatedDoorbots {
			// Only mark online if the doorbot has a waiting post request
			if door.ID == service.ID {
				if service.State != database.WorkingState {
					service.State = database.WorkingState
					service.Since = time.Now()
				}

				service.Name = door.Name

				foundDoorbots = append(foundDoorbots, service)
				updatedDoorbots = append(updatedDoorbots[:i], updatedDoorbots[i+1:]...)
				found = true
				break
			}
		}

		if !found {
			if service.State != database.BrokenState {
				service.State = database.BrokenState
				service.Since = time.Now()
			}

			foundDoorbots = append(foundDoorbots, service)
		}
	}

	for i, doorbot := range updatedDoorbots {
		service := &database.Service{
			ID:          doorbot.ID,
			Name:        doorbot.Name,
			State:       database.WorkingState,
			Since:       time.Now(),
			ServiceName: "doorbot",
		}

		foundDoorbots = append(foundDoorbots, service)
		updatedDoorbots = append(updatedDoorbots[:i], updatedDoorbots[i+1:]...)
	}

	return foundDoorbots, nil
}

func getExistingDoorbots(store *database.Store) ([]*database.Service, error) {
	services, err := store.Service.GetServices()

	if err != nil {
		return nil, err
	}

	var doorbots []*database.Service

	for _, door := range services {
		if door.ServiceName == "doorbot" {
			doorbots = append(doorbots, door)
		}
	}

	return doorbots, nil
}

// RecievePing is called by the api to mark a doorbot request as recieved
func RecievePing(doorbot *Doorbot) error {
	doorbot.Name = "Doorbot: " + strconv.FormatInt(doorbot.ID, 10)

	updatedDoorbots = append(updatedDoorbots, doorbot)

	return nil
}
