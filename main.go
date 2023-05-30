package main

import (
	"database/sql"
	"fmt"
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
	// Read the single rental from the database
	rentalID := c.Query("id")
	var rental Rental

	query := `
		SELECT r.id, r.name, r.description, r.type, r.make, r.model, r.year, r.length, r.sleeps, r.primary_image_url, r.price_day,
		l.city, l.state, l.zip, l.country, l.lat, l.lng,
		u.id, u.first_name, u.last_name
		FROM rentals r
		JOIN users u ON r.user_id = u.id
		WHERE r.id = $1
	`

	row := db.QueryRow(query, rentalID)
	err := row.Scan(&rental.ID, &rental.Name, &rental.Description, &rental.Type, &rental.Make, &rental.Model, &rental.Year, &rental.Length, &rental.Sleeps, &rental.PrimaryImageURL, &rental.Price.Day,
		&rental.Location.City, &rental.Location.State, &rental.Location.Zip, &rental.Location.Country, &rental.Location.Lat, &rental.Location.Lng,
		&rental.User.ID, &rental.User.FirstName, &rental.User.LastName)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("Rental not found")
		} else {
			log.Fatal(err)
		}
	}

	c.JSON(200, rental)
}

func getRentals(c *gin.Context, db *sql.DB) {
	rentals := []Rental{}

	query := `
		SELECT r.id, r.name, r.description, r.type, r.make, r.model, r.year, r.length, r.sleeps, r.primary_image_url, r.price_day,
		l.city, l.state, l.zip, l.country, l.lat, l.lng,
		u.id, u.first_name, u.last_name
		FROM rentals r
		JOIN users u ON r.user_id = u.id
	`

	rentalRows, err := db.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer rentalRows.Close()

	for rentalRows.Next() {
		var rental Rental
		err := rentalRows.Scan(&rental.ID, &rental.Name, &rental.Description, &rental.Type, &rental.Make, &rental.Model, &rental.Year, &rental.Length, &rental.Sleeps, &rental.PrimaryImageURL, &rental.Price.Day,
			&rental.Location.City, &rental.Location.State, &rental.Location.Zip, &rental.Location.Country, &rental.Location.Lat, &rental.Location.Lng,
			&rental.User.ID, &rental.User.FirstName, &rental.User.LastName)
		if err != nil {
			log.Fatal(err)
		}
		rentals = append(rentals, rental)
	}

	if err != nil {
		log.Fatal(err)
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
