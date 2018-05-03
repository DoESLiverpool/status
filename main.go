package main

import (
	"fmt"
	"net/http"

	"github.com/caarlos0/env"

	"github.com/DoESLiverpool/status/database"
	"github.com/DoESLiverpool/status/services"

	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
)

// SystemSettings are the required settings for the system to know how to run
type SystemSettings struct {
	Port string `env:"HTTP_PORT"`
	Mode string `env:"GIN_MODE"`
}

func main() {
	var settings = SystemSettings{}
	env.Parse(&settings)

	fmt.Println("Port set to:", settings.Port)
	fmt.Println("Running in:", settings.Mode)

	if gin.ReleaseMode == settings.Mode {
		gin.SetMode(gin.ReleaseMode)
	}

	fmt.Println("Updating data")
	err := updateData()

	if err != nil {
		fmt.Println("Error updating data")
		panic(err)
		return
	}

	router := gin.Default()

	router.Use(static.Serve("/", static.LocalFile("./public", true)))

	api := router.Group("/api")

	// git := api.Group("/git")
	// {
	// git.GET("/labels", func(c *gin.Context) {
	// c.JSON(http.StatusOK, labels)
	// })

	// git.GET("/issues", func(c *gin.Context) {
	// c.JSON(http.StatusOK, issues)
	// })

	// git.GET("/services", func(c *gin.Context) {
	// c.JSON(http.StatusOK, githubServices)
	// })
	// }

	api.GET("/status", func(c *gin.Context) {
		store := database.Store{}

		err := store.GetDatabase(true)
		defer store.CloseDatabase()

		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
		} else {

			services, err := store.GetServices()

			if err != nil {
				c.JSON(http.StatusInternalServerError, err)
			} else {
				c.JSON(http.StatusOK, services)
			}
		}
	})

	router.Run(settings.Port)
}

func updateData() error {
	githubServices, err := services.UpdateGit()

	if err != nil {
		return err
	}

	store := database.Store{}
	err = store.GetDatabase(false)

	if err != nil {
		return err
	}

	for _, githubService := range githubServices {
		err := store.CreateService(githubService)

		if err != nil {
			return err
		}
	}

	store.CloseDatabase()

	return nil
}
