package main

import (
	"bytes"
	"log"
	"net/http"

	"github.com/numyalai/fog-computing-assignment/pkg/util"
)

func main() {
	log.SetPrefix("router: ")
	log.Println("Starting ...")

	listenAddr := "localhost:5001"
	server := http.NewServeMux()

	buf := make([]string, 0)
	var reqBuffer = util.RouterRequestBuffer{Buffer: &buf}
	go util.RouterSendLoop(&reqBuffer, "http://localhost:5002/forward")

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

	server.HandleFunc("/forward", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received %s request from %s", r.Method, r.RemoteAddr)
		buffer := new(bytes.Buffer)
		buffer.ReadFrom(r.Body)
		body := buffer.String()
		log.Println(body)
		reqBuffer.Mu.Lock()
		*reqBuffer.Buffer = append(*reqBuffer.Buffer, body)
		reqBuffer.Mu.Unlock()
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK\n"))
	})

	log.Printf("Serving at %s", listenAddr)
	err := http.ListenAndServe(listenAddr, server)

	if err != nil {
		log.Printf("%s", err)
	}

	log.Println("Stopped.")
}
