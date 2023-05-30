package main

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"

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
		SELECT r.*, u.id as user_id, u.first_name as user_first_name, u.last_name as user_last_name
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
	// Get the list of IDs from the query parameter
	ids := c.Query("ids")
	sort := c.Query("sort")
	priceMin := c.Query("price_min")
	priceMax := c.Query("price_max")
	near := c.Query("near")
	offset := c.DefaultQuery("offset", "0")
	limit := c.DefaultQuery("limit", "10")

	var query string
	var values []interface{}

	if ids != "" {
		// Split the IDs into a slice
		idSlice := strings.Split(ids, ",")

		// Construct the placeholders for the parameterized query
		placeholders := make([]string, len(idSlice))
		values = make([]interface{}, len(idSlice))
		for i, id := range idSlice {
			placeholders[i] = "$" + strconv.Itoa(i+1)
			values[i] = id
		}

		// Construct the SQL query with the parameterized query
		query = "SELECT r.*, u.id as user_id, u.first_name as user_first_name, u.last_name as user_last_name FROM rentals r JOIN users u ON r.user_id = u.id  WHERE r.id IN (" + strings.Join(placeholders, ",") + ")"
	} else {
		// No filtering required, fetch all rentals
		query = "SELECT r.*, u.id as user_id, u.first_name as user_first_name, u.last_name as user_last_name FROM rentals JOIN users u ON r.user_id = u.id"
	}

	if priceMin != "" {
		// Add the price_min filter to the query
		priceMinValue, err := strconv.ParseFloat(priceMin, 64)
		if err != nil {
			// Handle the error
			c.JSON(400, gin.H{"error": "Invalid price_min value"})
			return
		}
		query += " AND r.price >= $" + strconv.Itoa(len(values)+1)
		values = append(values, priceMinValue)
	}

	if priceMax != "" {
		// Add the price_max filter to the query
		priceMaxValue, err := strconv.ParseFloat(priceMax, 64)
		if err != nil {
			// Handle the error
			c.JSON(400, gin.H{"error": "Invalid price_max value"})
			return
		}
		query += " AND r.price <= $" + strconv.Itoa(len(values)+1)
		values = append(values, priceMaxValue)
	}

	if near != "" {
		// Split the latLng into latitude and longitude
		latLng := strings.Split(near, ",")
		if len(latLng) != 2 {
			// Handle the error
			c.JSON(400, gin.H{"error": "Invalid near value"})
			return
		}

		// Convert the latitude and longitude to float64
		lat, err := strconv.ParseFloat(latLng[0], 64)
		if err != nil {
			// Handle the error
			c.JSON(400, gin.H{"error": "Invalid latitude value"})
			return
		}

		lng, err := strconv.ParseFloat(latLng[1], 64)
		if err != nil {
			// Handle the error
			c.JSON(400, gin.H{"error": "Invalid longitude value"})
			return
		}

		// Add the near filter to the query
		query += " AND earth_box(ll_to_earth($" + strconv.Itoa(len(values)+1) + ", $" + strconv.Itoa(len(values)+2) + "), 100 * 1609.34) @> ll_to_earth(r.lat, r.lng)"
		values = append(values, lat, lng)
	}

	if sort != "" {
		// Add the sort parameter to the query
		sortField := ""
		switch sort {
		case "price":
			sortField = "r.price"
		case "id":
			sortField = "r.id"
		case "name":
			sortField = "r.name"
		case "description":
			sortField = "r.description"
		case "type":
			sortField = "r.type"
		case "make":
			sortField = "r.make"
		case "model":
			sortField = "r.model"
		case "year":
			sortField = "r.year"
		case "length":
			sortField = "r.length"
		case "sleeps":
			sortField = "r.sleeps"
		case "primary_image_url":
			sortField = "r.primary_image_url"
		case "lat":
			sortField = "r.lat"
		case "lng":
			sortField = "r.lng"
		case "user_id":
			sortField = "u.id"
		case "user_first_name":
			sortField = "u.first_name"
		case "user_last_name":
			sortField = "u.last_name"
		default:
			c.JSON(400, gin.H{"error": "Invalid sort parameter"})
			return
		}

		query += " ORDER BY " + sortField
	}

	// Add the limit filter to the query
	limitValue, err := strconv.Atoi(limit)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid limit value"})
		return
	}
	query += " LIMIT $" + strconv.Itoa(len(values)+1)
	values = append(values, limitValue)

	// Add the offset filter to the query
	offsetValue, err := strconv.Atoi(offset)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid offset value"})
		return
	}
	query += " OFFSET $" + strconv.Itoa(len(values)+1)
	values = append(values, offsetValue)

	// Execute the SQL query with the provided values
	rows, err := db.Query(query, values...)
	if err != nil {
		// Handle the error
		c.JSON(500, gin.H{"error": "Failed to retrieve rentals"})
		return
	}
	defer rows.Close()

	// Iterate through the result set and build the rental list
	rentals := make([]Rental, 0)
	for rows.Next() {
		rental := Rental{}
		err := rows.Scan(
			&rental.ID,
			&rental.Name,
			&rental.Description,
			&rental.Type,
			&rental.Make,
			&rental.Model,
			&rental.Year,
			&rental.Length,
			&rental.Sleeps,
			&rental.PrimaryImageURL,
			&rental.Price,
			&rental.Location.Lat,
			&rental.Location.Lng,
			&rental.User.ID,
			&rental.User.FirstName,
			&rental.User.LastName,
		)
		if err != nil {
			// Handle the error
			c.JSON(500, gin.H{"error": "Failed to retrieve rentals"})
			return
		}
		rentals = append(rentals, rental)
	}

	// Return the filtered rentals as JSON response
	c.JSON(200, rentals)
}

func main() {
	// Connect to the PostgreSQL database
	db, err := sql.Open("pgx", "postgres://root:root@localhost:5432/testingwithrentals")
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
