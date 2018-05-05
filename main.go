package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/caarlos0/env"

	"github.com/DoESLiverpool/status/database"
	"github.com/DoESLiverpool/status/services"

	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
)

// SystemSettings are the required settings for the system to know how to run
type SystemSettings struct {
	Port         string `env:"HTTP_PORT"`
	Mode         string `env:"GIN_MODE"`
	UpdateTimer  int    `env:"UPDATE_TIME"`
	DoorbotToken string `env:"DOORBOT_API_KEY"`
}

var lastUpdated time.Time

func main() {
	var settings = SystemSettings{}
	env.Parse(&settings)

	fmt.Println("Port set to:", settings.Port)
	fmt.Println("Running in:", settings.Mode)

	if gin.ReleaseMode == settings.Mode {
		gin.SetMode(gin.ReleaseMode)
	}

	// Run the data fetcher in the background
	go dataUpdater(settings.UpdateTimer)

	router := gin.Default()

	router.Use(static.Serve("/", static.LocalFile("./public", true)))

	api := router.Group("/api")

	api.POST("/doorbot", func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		if authHeader != "Bearer "+settings.DoorbotToken {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		var doorbot services.Doorbot
		c.BindJSON(&doorbot)
		services.RecievePing(&doorbot)

		c.JSON(http.StatusOK, doorbot)
	})

	api.GET("/status", func(c *gin.Context) {
		store := database.Store{}

		err := store.GetDatabase(true)
		defer store.CloseDatabase()

		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
		} else {

			services, err := store.Service.GetServices()

			if err != nil {
				c.JSON(http.StatusInternalServerError, err)
			} else {
				var response struct {
					Updated  time.Time           `json:"updated"`
					Services []*database.Service `json:"services"`
				}

				response.Updated = lastUpdated
				response.Services = services

				c.JSON(http.StatusOK, response)
			}
		}
	})

	router.Run(settings.Port)
}

func dataUpdater(updateTimer int) {
	for {
		fmt.Println("Updating data")
		updateData()

		lastUpdated = time.Now()

		time.Sleep(time.Duration(updateTimer) * time.Second)
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
	defer store.CloseDatabase()

	if err != nil {
		return err
	}

	// Ensure git isn't updated whilst it is disabled
	if githubServices != nil {
		err = store.Service.UpdateServices(githubServices)

		if err != nil {
			return err
		}
	}

	// Ensure doorbots aren't updated whilst disabled
	if doorbotServices != nil {
		err = store.Service.UpdateServices(doorbotServices)

		if err != nil {
			return err
		}
	}

	return nil
}
