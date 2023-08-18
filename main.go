package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-sql-driver/mysql"

	"github.com/sadnessOjisan/gochann/router"
)

func main() {
	config := &mysql.Config{
		User:                 os.Getenv("SADNESS_MYSQL_USER"),
		Passwd:               os.Getenv("SADNESS_MYSQL_PASSWORD"),
		Net:                  "tcp",
		Addr:                 os.Getenv("SADNESS_MYSQL_HOST"),
		DBName:               os.Getenv("SADNESS_MYSQL_DATABASE"),
		ParseTime:            true,
		AllowNativePasswords: true,
	}
	db, err := sql.Open("mysql", config.FormatDSN())
	if err != nil {
		log.Fatalf("ERROR: db open err: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Fatalf("ERROR: db close err: %v", err)
		}
	}()

	h := router.NewHandler(db)

	mux := http.NewServeMux()
	mux.HandleFunc("/", h.HomeHandler)
	// for /users/:id
	mux.HandleFunc("/users", h.UsersHandler)
	mux.HandleFunc("/users/", h.UsersDetailHandler)

	mux.HandleFunc("/posts", h.PostsHandler)
	mux.HandleFunc("/posts/", h.PostsDetailHandler)
	mux.HandleFunc("/posts/new", h.PostsNewHandler)

	mux.HandleFunc("/signout", h.SignoutHandler)

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: 30 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}
