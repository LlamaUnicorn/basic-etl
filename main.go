package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type Comment struct {
	PostID int    `json:"postId"`
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Body   string `json:"body"`
}

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Construct the Data Source Name (DSN)
	dsn := fmt.Sprintf(
		"host=localhost port=%s dbname=%s user=%s password=%s sslmode=disable",
		os.Getenv("PG_PORT"),
		os.Getenv("PG_DATABASE_NAME"),
		os.Getenv("PG_USER"),
		os.Getenv("PG_PASSWORD"),
	)

	// Open a connection to the database
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Error opening database connection: %v", err)
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			log.Fatalf("Error closing database connection: %v", err)
		}
	}(db)

	// Verify the connection to the database
	if err := db.Ping(); err != nil {
		log.Fatalf("Error pinging database: %v", err)
	}

	// Define batch size and starting point
	batchSize := 50
	start := 0

	for {
		// Construct the API URL with pagination parameters
		url := fmt.Sprintf("https://jsonplaceholder.typicode.com/comments?_start=%d&_limit=%d", start, batchSize)

		// Make the HTTP GET request
		resp, err := http.Get(url)
		if err != nil {
			log.Fatalf("Error fetching data from API: %v", err)
		}

		// Ensure the response body is closed after processing
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				log.Fatalf("Error closing response body: %v", err)
			}
		}(resp.Body)

		// Check for a successful response
		if resp.StatusCode != http.StatusOK {
			log.Fatalf("Unexpected status code: %d", resp.StatusCode)
		}

		// Decode the JSON response into a slice of Comment structs
		var comments []Comment
		if err := json.NewDecoder(resp.Body).Decode(&comments); err != nil {
			log.Fatalf("Error decoding JSON response: %v", err)
		}

		// Break the loop if no more comments are returned
		if len(comments) == 0 {
			log.Println("No more comments to process. Exiting loop.")
			break
		}

		// Begin a new transaction
		tx, err := db.Begin()
		if err != nil {
			log.Fatalf("Error starting transaction: %v", err)
		}

		// Build the SQL insert statement using squirrel
		qb := squirrel.Insert("comments").
			Columns("post_id", "id", "name", "email", "body").
			PlaceholderFormat(squirrel.Dollar)

		// Add values to the insert statement
		for _, comment := range comments {
			qb = qb.Values(comment.PostID, comment.ID, comment.Name, comment.Email, comment.Body)
		}

		// Generate the SQL and arguments
		sqlStr, args, err := qb.ToSql()
		if err != nil {
			err := tx.Rollback()
			if err != nil {
				return
			}
			log.Fatalf("Error building SQL statement: %v", err)
		}

		// Execute the insert statement
		if _, err := tx.Exec(sqlStr, args...); err != nil {
			err := tx.Rollback()
			if err != nil {
				return
			}
			log.Fatalf("Error executing insert statement: %v", err)
		}

		// Commit the transaction
		if err := tx.Commit(); err != nil {
			log.Fatalf("Error committing transaction: %v", err)
		}

		log.Printf("Successfully inserted batch starting at %d with %d comments", start, len(comments))

		// Increment the starting point for the next batch
		start += batchSize
		time.Sleep(1 * time.Second)
	}
}

// TODO: add validation
