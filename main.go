package main

import (
	"database/sql"
	"log"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v4/stdlib"
)

type Rental struct {
	ID              int     `json:"id"`
	Name            string  `json:"name"`
	Description     string  `json:"description"`
	Type            string  `json:"type"`
	Make            string  `json:"make"`
	Model           string  `json:"model"`
	Year            int     `json:"year"`
	Length          float64 `json:"length"`
	Sleeps          int     `json:"sleeps"`
	PrimaryImageURL string  `json:"primary_image_url"`
	Price           struct {
		Day int `json:"day"`
	} `json:"price"`
	Location struct {
		City    string  `json:"city"`
		State   string  `json:"state"`
		Zip     string  `json:"zip"`
		Country string  `json:"country"`
		Lat     float64 `json:"lat"`
		Lng     float64 `json:"lng"`
	} `json:"location"`
	User struct {
		ID        int    `json:"id"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	} `json:"user"`
}

func getRental(c *gin.Context, db *sql.DB) {
	rental := Rental{
		ID:       1,
		Name:     "Sample Rental",
		Price: struct {
			Day int `json:"day"`
		}{
			Day: 100,
		},
		Location: struct {
			City  string `json:"city"`
			State string `json:"state"`
		}{
			City:  "Sample City",
			State: "Sample State",
		},
	}

	}

	c.JSON(200, rental)
}

func getRentals(c *gin.Context, db *sql.DB) {
	// Logic to retrieve and filter rentals based on the query parameters
	// Use the db connection to fetch data from the PostgreSQL database
	// Transform the data into a list of rental objects in the JSON structure
	rentals := []Rental{
		{
			ID:        1,
			Name:      "Sample Rental 1",
			Price: struct {
				Day int `json:"day"`
			}{
				Day: 100,
			},
			Location: struct {
				City  string `json:"city"`
				State string `json:"state"`
			}{
				City:  "Sample City 1",
				State: "Sample State 1",
			},
		},
		{
			ID:        2,
			Name:      "Sample Rental 2",
			Price: struct {
				Day int `json:"day"`
			}{
				Day: 200,
			},
			Location: struct {
				City  string `json:"city"`
				State string `json:"state"`
			}{
				City:  "Sample City 2",
				State: "Sample State 2",
			},
		},
	}


	c.JSON(200, rentals)
}

func main() {
	// Connect to the PostgreSQL database
	db, err := sql.Open("pgx", "postgres://user:password@localhost:5432/database")
	if err != nil {
		log.Fatal("Failed to connect to the database:", err)
	}
	defer db.Close()

	// Initialize the gin engine
	router := gin.Default()

	// Define the routes and handlers
	router.GET("/rentals/:rentalID", func(c *gin.Context) {
		getRental(c, db)
	})
	router.GET("/rentals", func(c *gin.Context) {
		getRentals(c, db)
	})

	// Run the application
	err = router.Run(":8080")
	if err != nil {
		log.Fatal("Failed to start the server:", err)
	}
}
