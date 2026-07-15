package main

import (
	"fmt"
	"net/http"
)

func main() {
	fs := http.FileServer(http.Dir("./dashboard"))
	http.Handle("/", fs)

	fmt.Println("Serving MEV HUD on http://localhost:8080 ...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
