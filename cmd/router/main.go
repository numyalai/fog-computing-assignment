package main

import (
	"bytes"
	"log"
	"net/http"
)

func main() {
	log.SetPrefix("router: ")
	log.Println("Starting ...")

	listenAddr := "localhost:5001"
	server := http.NewServeMux()

	server.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received %s request from %s", r.Method, r.RemoteAddr)
		log.Println(r)
		buffer := new(bytes.Buffer)
		buffer.ReadFrom(r.Body)
		body := buffer.String()
		log.Println("Body := " + string(body))
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	server.HandleFunc("/forward/", func(w http.ResponseWriter, r *http.Request) {
		// TODO: enable forwarding requests from here to be adaptable to any software running behind this providing HTTP endpoints
	})

	log.Printf("Serving at %s", listenAddr)
	err := http.ListenAndServe(listenAddr, server)

	if err != nil {
		log.Printf("%s", err)
	}

	log.Println("Stopped.")
}
