package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

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
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	dsn := fmt.Sprintf(
		"host=localhost port=%s dbname=%s user=%s password=%s sslmode=disable",
		os.Getenv("PG_PORT"),
		os.Getenv("PG_DATABASE_NAME"),
		os.Getenv("PG_USER"),
		os.Getenv("PG_PASSWORD"),
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Error opening database connection: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Error pinging database: %v", err)
	}

	for i := 0; i < 5; i++ {
		start := i * 100
		url := fmt.Sprintf("https://jsonplaceholder.typicode.com/comments?_start=%d&_limit=100", start)
		resp, err := http.Get(url)
		if err != nil {
			log.Fatalf("Error fetching batch %d: %v", i, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Fatalf("Unexpected status code %d for batch %d", resp.StatusCode, i)
		}

		var comments []Comment
		if err := json.NewDecoder(resp.Body).Decode(&comments); err != nil {
			log.Fatalf("Error decoding batch %d: %v", i, err)
		}

		tx, err := db.Begin()
		if err != nil {
			log.Fatalf("Error starting transaction for batch %d: %v", i, err)
		}

		qb := squirrel.Insert("comments").
			Columns("post_id", "id", "name", "email", "body").
			PlaceholderFormat(squirrel.Dollar)

		for _, comment := range comments {
			qb = qb.Values(comment.PostID, comment.ID, comment.Name, comment.Email, comment.Body)
		}

		sql, args, err := qb.ToSql()
		if err != nil {
			tx.Rollback()
			log.Fatalf("Error building SQL for batch %d: %v", i, err)
		}

		if _, err := tx.Exec(sql, args...); err != nil {
			tx.Rollback()
			log.Fatalf("Error inserting batch %d: %v", i, err)
		}

		if err := tx.Commit(); err != nil {
			log.Fatalf("Error committing transaction for batch %d: %v", i, err)
		}

		log.Printf("Successfully inserted batch %d with %d comments", i, len(comments))
	}
}
