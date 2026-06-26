package main

import (
	"fmt"
	"net/http"
)

func main() {
	//1.handler untuk path "/"
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello from Go")
	})

	//2. handler untuk path "/healthz"
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "OK")
	})

	//3. run server di port 8080
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Waduh error:", err)
	}
}
