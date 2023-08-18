package main

import (
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/sadnessOjisan/gochann/router"
)

func main() {
	http.HandleFunc("/", router.HomeHandler)
	// for /users/:id
	http.HandleFunc("/users", router.UsersHandler)
	http.HandleFunc("/users/", router.UsersDetailHandler)

	http.HandleFunc("/posts", router.PostsHandler)
	http.HandleFunc("/posts/", router.PostsDetailHandler)
	http.HandleFunc("/posts/new", router.PostsNewHandler)

	http.HandleFunc("/signout", router.SignoutHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
