package main

import (
	"log"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"github.com/sadnessOjisan/gochann/router"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", router.HomeHandler)
	// for /users/:id
	mux.HandleFunc("/users", router.UsersHandler)
	mux.HandleFunc("/users/", router.UsersDetailHandler)

	mux.HandleFunc("/posts", router.PostsHandler)
	mux.HandleFunc("/posts/", router.PostsDetailHandler)
	mux.HandleFunc("/posts/new", router.PostsNewHandler)

	mux.HandleFunc("/signout", router.SignoutHandler)

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: 30 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}
