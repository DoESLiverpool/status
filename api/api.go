package api

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/DoESLiverpool/status/database"
	"github.com/DoESLiverpool/status/services"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
)

// Handlers n
type Handlers struct {
	Router   *gin.Engine
	Settings services.SystemSettings
}

// LoadRoutes will generate all the api handlers
func (h *Handlers) LoadRoutes(settings services.SystemSettings) {
	h.Settings = settings

	h.Router = gin.Default()

	h.Router.Use(static.Serve("/", static.LocalFile("./public", true)))

	api := h.Router.Group("/api")

	api.POST("/doorbot", h.doorbotUpdate)
	api.GET("/status", h.status)
	api.GET("/history/:service", h.historyForService)
}

// StartListening will pull the port from settings and begin listening
func (h Handlers) StartListening() {
	fmt.Println("Port set to:", h.Settings.Port)
	h.Router.Run(h.Settings.Port)
}

func (h Handlers) doorbotUpdate(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")

	if authHeader != "Bearer "+h.Settings.DoorbotToken {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	var doorbot services.Doorbot
	c.BindJSON(&doorbot)
	services.RecievePing(&doorbot)

	c.JSON(http.StatusOK, doorbot)
}

func (h Handlers) status(c *gin.Context) {
	store := database.Store{}

	err := store.GetDatabase(true)
	defer store.CloseDatabase()

	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
	} else {

		srvs, err := store.Service.GetServices()

		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
		} else {
			var response struct {
				Updated  time.Time           `json:"updated"`
				Services []*database.Service `json:"services"`
			}

			response.Updated = services.LastUpdatedTime
			response.Services = srvs

			c.JSON(http.StatusOK, response)
		}
	}
}

func (h Handlers) historyForService(c *gin.Context) {
	service := c.Param("service")
	serviceID, err := strconv.ParseInt(service, 10, 64)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid service id",
		})

		return
	}

	store := database.Store{}

	err = store.GetDatabase(true)
	defer store.CloseDatabase()

	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
	} else {
		history, err := store.History.GetHistoryForService(serviceID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
		} else {
			c.JSON(http.StatusOK, history)
		}
	}
}
