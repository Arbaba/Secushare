package main

import (
    "fmt"
    "net/http"
    "github.com/gorilla/mux"
)

func main() {
    r := mux.NewRouter()
    r.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))
 
	
	
    r.HandleFunc("/messages/list", func(w http.ResponseWriter, r *http.Request) {
        
        fmt.Fprintf(w, "")
	})

	r.HandleFunc("/messages/send", func(w http.ResponseWriter, r *http.Request) {
        
        fmt.Fprintf(w, "")
	})
	
	

	
    r.HandleFunc("/peers/list", func(w http.ResponseWriter, r *http.Request) {
        
        fmt.Fprintf(w, "")
	})
	
	
    r.HandleFunc("/peers/add", func(w http.ResponseWriter, r *http.Request) {
        
        fmt.Fprintf(w, "")
    })

    http.ListenAndServe(":8080", r)
}