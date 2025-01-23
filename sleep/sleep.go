package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"
)

func helloWorldHandler(w http.ResponseWriter, r *http.Request) {

	delay := time.Duration(rand.Intn(5)+10) * time.Millisecond
	time.Sleep(delay)

	fmt.Fprintln(w, "Hello, world!")
}

func main() {

	rand.Seed(time.Now().UnixNano())

	http.HandleFunc("/", helloWorldHandler)

	log.Println("Servidor HTTP iniciado en :8096")
	log.Fatal(http.ListenAndServe(":8096", nil))
}
