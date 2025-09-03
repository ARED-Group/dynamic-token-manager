package main

import (
	"fmt"
	"log"
	"net/http"
	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Dynamic Token Manager is running!")
	})
	fmt.Println("Server starting at :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}