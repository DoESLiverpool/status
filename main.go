package main

import (
	"fmt"
	"time"

	"github.com/DoESLiverpool/status/api"
	"github.com/DoESLiverpool/status/database"
	"github.com/DoESLiverpool/status/services"

	"github.com/gin-gonic/gin"
)

var handler api.Handlers

func main() {
	settings := services.GetSystemSettings()

	fmt.Println("Port set to:", settings.Port)
	fmt.Println("Running in:", settings.Mode)

	if gin.ReleaseMode == settings.Mode {
		gin.SetMode(gin.ReleaseMode)
	}

	// Run the data fetcher in the background
	go dataUpdater(settings.UpdateTimer)

	handler := &api.Handlers{}

	// Load all the routes
	handler.LoadRoutes(settings)

	// Begin listen
	handler.StartListening()
}

func dataUpdater(updateTimer int) {
	for {
		fmt.Println("Updating data")
		err := updateData()

		if err != nil {
			fmt.Printf(err.Error())
		} else {
			services.LastUpdatedTime = time.Now()

			time.Sleep(time.Duration(updateTimer) * time.Second)
		}
	}
}

func updateData() error {
	githubServices, err := services.UpdateGit()

	if err != nil {
		return err
	}

	doorbotServices, err := services.UpdateDoorbots()

	if err != nil {
		return err
	}

	store := database.Store{}
	err = store.GetDatabase(false)

	if err != nil {
		store.CloseDatabase()
		return err
	}

	// Ensure git isn't updated whilst it is disabled
	if githubServices != nil {
		err = store.Service.UpdateServices(githubServices)

		if err != nil {
			store.CloseDatabase()
			return err
		}
	}

	// Ensure doorbots aren't updated whilst disabled
	if doorbotServices != nil {
		err = store.Service.UpdateServices(doorbotServices)

		if err != nil {
			store.CloseDatabase()
			return err
		}
	}

	store.CloseDatabase()
	return nil
}
