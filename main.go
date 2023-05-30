package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Rental struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func main() {
	router := gin.Default()
	router.GET("/rentals/:id", getRental)
	router.GET("/rentals", getRentals)
	router.Run(":8080")
}

func getRental(c *gin.Context) {
	rentalID := c.Param("id")
	rental := Rental{ID: rentalID, Name: "Sample Rental"}
	c.JSON(http.StatusOK, rental)
}

func getRentals(c *gin.Context) {
	//location := c.Query("location")
	rentals := []Rental{{ID: "1", Name: "Rental 1"}, {ID: "2", Name: "Rental 2"}}
	c.JSON(http.StatusOK, rentals)
}
