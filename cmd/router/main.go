package main

import (
	"bytes"
	"log"
	"net/http"
	"time"

	"github.com/numyalai/fog-computing-assignment/pkg/util"
)

func deregisterInactiveClients(clientStorage *util.Storage) {
	for {
		clientStorage.DeregisterInactiveClients(5 * time.Minute)
		time.Sleep(5 * time.Minute)
	}
}

func main() {
	log.SetPrefix("router: ")
	log.Println("Starting ...")

	listenAddr := "localhost:5001"
	server := http.NewServeMux()

	storage := util.NewStorage()

	buf := make([]string, 0)
	var reqBuffer = util.RequestBuffer{Buffer: &buf}
	go util.SendLoop(&reqBuffer, "http://localhost:5002/forward")

	go deregisterInactiveClients(storage)

	server.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received %s request from %s", r.Method, r.RemoteAddr)
		log.Println(r)
		buffer := new(bytes.Buffer)
		buffer.ReadFrom(r.Body)
		body := buffer.String()
		cliendID := r.Header.Get("Client-ID")
		storage.RegisterClient(cliendID, "1GB", "1GHz")
		storage.UpdateClient(cliendID, "2GB", "2GHz")
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
