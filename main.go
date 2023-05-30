package main

import (
	"database/sql"
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
		Day float64 `json:"day"`
	} `json:"price_per_day"`
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
	rentalID := c.Param("id")
	var rental Rental

	query := `SELECT 
			r.id,
			r.name,
			r.description,
			r.type,
			r.vehicle_make,
			r.vehicle_model,
			r.vehicle_year,
			r.vehicle_length,
			r.sleeps,
			r.primary_image_url,
			r.price_per_day,
			r.home_city,
			r.home_state,
			r.home_zip,
			r.home_country,
			r.lat,
			r.lng,
			u.id as user_id,
			u.first_name as user_first_name,
			u.last_name as user_last_name
		FROM rentals r 
		JOIN users u 
			ON r.user_id = u.id 
		WHERE r.id = $1`

	row := db.QueryRow(query, rentalID)
	err := row.Scan(
		&rental.ID,               //1
		&rental.Name,             //2
		&rental.Description,      //3
		&rental.Type,             //4
		&rental.Make,             //5
		&rental.Model,            //6
		&rental.Year,             //7
		&rental.Length,           //8
		&rental.Sleeps,           //9
		&rental.PrimaryImageURL,  //10
		&rental.Price.Day,        //11
		&rental.Location.City,    //12
		&rental.Location.State,   //13
		&rental.Location.Zip,     //14
		&rental.Location.Country, //15
		&rental.Location.Lat,     //16
		&rental.Location.Lng,     //17
		&rental.User.ID,          //18
		&rental.User.FirstName,   //19
		&rental.User.LastName,    //20
	)
	/*
		r.id, 1
				r.name, 2
				r.description, 3
				r.type, 4
				r.vehicle_make, 5
				r.vehicle_model, 6
				r.vehicle_year, 7
				r.vehicle_length, 8
				r.sleeps, 9
				r.primary_image_url, 10
				r.price_per_day, 11
				r.home_city, 12
				r.home_state, 13
				r.home_zip, 14
				r.home_country, 15
				r.lat, 16
				r.lng, 17
				u.id as user_id, 18
				u.first_name as user_first_name, 19
				u.last_name as user_last_name 20
	*/
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(400, gin.H{"error": "Not Found"})
		} else {
			log.Println(err)
			c.JSON(500, gin.H{"error": "Internal Server Error"})
		}
		return
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
	query = `SELECT 
			r.id, 
			r.name, 
			r.description, 
			r.type, 
			r.vehicle_make, 
			r.vehicle_model, 
			r.vehicle_year, 
			r.vehicle_length, 
			r.sleeps, 
			r.primary_image_url, 
			r.price_per_day, 
			r.home_city, 
			r.home_state, 
			r.home_zip, 
			r.home_country, 
			r.lat, 
			r.lng, 
			u.id as user_id, 
			u.first_name as user_first_name, 
			u.last_name as user_last_name 
		FROM rentals r 
		JOIN users u 
			ON r.user_id = u.id
		WHERE 1=1`
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
		query = query + " AND r.id IN (" + strings.Join(placeholders, ",") + ") "
	}

	if priceMin != "" {
		// Add the price_min filter to the query
		priceMinValue, err := strconv.ParseFloat(priceMin, 64)
		if err != nil {
			// Handle the error
			c.JSON(400, gin.H{"error": "Invalid price_min value"})
			return
		}
		query += " AND r.price_per_day >= $" + strconv.Itoa(len(values)+1)
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
		query += " AND r.price_per_day <= $" + strconv.Itoa(len(values)+1)
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
			sortField = "r.price_per_day"
		case "id":
			sortField = "r.id"
		case "name":
			sortField = "r.name"
		case "description":
			sortField = "r.description"
		case "type":
			sortField = "r.type"
		case "make":
			sortField = "r.vehicle_make"
		case "model":
			sortField = "r.vehicle_model"
		case "year":
			sortField = "r.vehicle_year"
		case "length":
			sortField = "r.vehicle_length"
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
			&rental.ID,               //1
			&rental.Name,             //2
			&rental.Description,      //3
			&rental.Type,             //4
			&rental.Make,             //5
			&rental.Model,            //6
			&rental.Year,             //7
			&rental.Length,           //8
			&rental.Sleeps,           //9
			&rental.PrimaryImageURL,  //10
			&rental.Price.Day,        //11
			&rental.Location.City,    //12
			&rental.Location.State,   //13
			&rental.Location.Zip,     //14
			&rental.Location.Country, //15
			&rental.Location.Lat,     //16
			&rental.Location.Lng,     //17
			&rental.User.ID,          //18
			&rental.User.FirstName,   //19
			&rental.User.LastName,    //20
		)
		if err != nil {
			// Handle the error
			log.Println(err)
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
	db, err := sql.Open("pgx", "postgres://root:root@postgres:5432/testingwithrentals")
	if err != nil {
		log.Println("Failed to connect to database.")
		log.Fatal("Failed to connect to the database:", err)
	}
	log.Println("Connected to database.")
	defer db.Close()

	// Initialize the gin engine
	router := gin.Default()

	// Define the routes and handlers
	router.GET("/rentals/:id", func(c *gin.Context) {
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
