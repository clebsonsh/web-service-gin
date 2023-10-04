package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
)

var db *sql.DB

// album represents data about a record album.
type album struct {
	ID     int64   `json:"id"`
	Title  string  `json:"title"`
	Artist string  `json:"artist"`
	Price  float32 `json:"price"`
}

func main() {
	// Capture connection properties.
	cfg := mysql.Config{
		User:                 os.Getenv("DBUSER"),
		Passwd:               os.Getenv("DBPASS"),
		Net:                  "tcp",
		Addr:                 "127.0.0.1:3306",
		DBName:               "recordings",
		AllowNativePasswords: true,
	}
	// Get a database handle.
	var err error
	db, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}

	pingErr := db.Ping()
	if pingErr != nil {
		log.Fatal(pingErr)
	}
	fmt.Println("Connected!")

	router := gin.Default()
	router.GET("/albums", getAlbums)
	router.GET("/albums/:id", getAlbumByID)
	router.POST("/albums", postAlbums)

	router.Run("localhost:8080")
}

// getAlbums responds with the list of all albums as JSON.
func getAlbums(c *gin.Context) {
	// An album slice to hold data from returned rows.
	var albums []album

	// Query the database.
	rows, err := db.Query("SELECT * FROM album")
	if err != nil {
		fmt.Printf("error: %v", err)
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "error getting albums"})
		return
	}
	defer rows.Close()

	// Loop through rows, using Scan to assign column data to struct fields.
	for rows.Next() {
		var alb album
		if err := rows.Scan(&alb.ID, &alb.Title, &alb.Artist, &alb.Price); err != nil {
			fmt.Printf("error %q: ", err)
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "error getting albums"})
			return
		}
		albums = append(albums, alb)
	}

	// Check for errors encountered during iteration.
	if err := rows.Err(); err != nil {
		fmt.Printf("error: %v", err)
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "error getting albums"})
		return
	}

	c.IndentedJSON(http.StatusOK, albums)
}

// postAlbums adds an album from JSON received in the request body.
func postAlbums(c *gin.Context) {
	var album album

	// Call BindJSON to bind the received JSON to newAlbum.
	if err := c.BindJSON(&album); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "error adding album"})
		return
	}

	_, err := db.Exec("INSERT INTO album (title, artist, price) VALUES (?, ?, ?)", album.Title, album.Artist, album.Price)
	if err != nil {
		fmt.Printf("addAlbum: %v", err)
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "error adding album"})
		return
	}

	c.IndentedJSON(http.StatusCreated, nil)
}

// getAlbumByID locates the album whose ID value matches the id
// parameter sent by the client, then returns that album as a response.
func getAlbumByID(c *gin.Context) {
	id := c.Param("id")
	// An album to hold data from the returned row.
	var alb album

	// Query the database for a single row.
	row := db.QueryRow("SELECT * FROM album WHERE id = ?", id)

	// Unmarshal the row's data into the fields of an Album.
	if err := row.Scan(&alb.ID, &alb.Title, &alb.Artist, &alb.Price); err != nil {
		if err == sql.ErrNoRows {
			c.IndentedJSON(http.StatusNotFound, gin.H{"message": "album not found"})
			return
		}
	}

	c.IndentedJSON(http.StatusOK, alb)
}
