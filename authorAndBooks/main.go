package main

import (
	"authorAndBooks/database"
	"authorAndBooks/handler/author"
	"authorAndBooks/handler/book"
	"authorAndBooks/service"
	"authorAndBooks/store"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

const readHeaderTimeout = 5 * time.Second

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	db, err := database.Connect(database.ConfigFromEnv())
	if err != nil {
		return err
	}
	defer db.Close()

	authorStore := store.NewAuthor(db)
	bookStore := store.NewBook(db)
	tx := store.NewTransactor(db)

	authorHandler := author.New(service.NewAuthor(authorStore, tx))
	bookHandler := book.New(service.NewBook(bookStore, authorStore, tx))

	r := mux.NewRouter()

	r.HandleFunc("/book", bookHandler.GetAll).Methods(http.MethodGet)
	r.HandleFunc("/book/{id}", bookHandler.GetByID).Methods(http.MethodGet)
	r.HandleFunc("/book", bookHandler.Create).Methods(http.MethodPost)
	r.HandleFunc("/book/{id}", bookHandler.Update).Methods(http.MethodPut)
	r.HandleFunc("/book/{id}", bookHandler.Delete).Methods(http.MethodDelete)
	r.HandleFunc("/author", authorHandler.Create).Methods(http.MethodPost)
	r.HandleFunc("/author/{id}", authorHandler.Update).Methods(http.MethodPut)
	r.HandleFunc("/author/{id}", authorHandler.Delete).Methods(http.MethodDelete)

	server := http.Server{
		Addr:              ":8000",
		Handler:           r,
		ReadHeaderTimeout: readHeaderTimeout,
	}

	log.Println("server started at :8000")

	return server.ListenAndServe()
}
