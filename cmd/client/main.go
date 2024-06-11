package main

import (
	"bytes"
	"log"
	"net/http"
	"time"
)

var baseSleep = 1000
var sleepFactor = 1

func sendLoop(reqBuffer *RequestBuffer) {
	client := &http.Client{}
	for {
		if len(*reqBuffer.buffer) <= 0 {
			var sleepDuration = baseSleep * sleepFactor
			time.Sleep(time.Duration(sleepDuration) * time.Millisecond)
			if sleepFactor < 120 {
				sleepFactor = sleepFactor * 2
			}
			continue
		}
		reqBuffer.mu.Lock()
		req := (*reqBuffer.buffer)[0]
		log.Println("Request := ", req)
		tmp, err := http.NewRequest("POST", "http://localhost:5001", bytes.NewBufferString(req))

		if err != nil {
			reqBuffer.mu.Unlock()
			log.Println("Unable to create HTTP POST request", err)
			return
		}

		resp, err := client.Do(tmp)

		if err != nil {
			var sleepDuration = baseSleep * sleepFactor
			time.Sleep(time.Duration(sleepDuration) * time.Millisecond)
			if sleepFactor < 120 {
				sleepFactor = sleepFactor * 2
			}
			reqBuffer.mu.Unlock()
			continue
		}

		log.Println(resp)

		*reqBuffer.buffer = (*reqBuffer.buffer)[1:] // remove handled element from queue
		reqBuffer.mu.Unlock()
	}
}

func main() {
	log.SetPrefix("client: ")
	log.Println("Starting ...")

	listenAddr := "localhost:5002"
	server := http.NewServeMux()

	buf := make([]string, 0)
	var reqBuffer = RequestBuffer{buffer: &buf}
	go sendLoop(&reqBuffer)

	server.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received %s request from %s", r.Method, r.RemoteAddr)
		buffer := new(bytes.Buffer)
		buffer.ReadFrom(r.Body)
		body := buffer.String()
		log.Println(body)
		reqBuffer.mu.Lock()
		*reqBuffer.buffer = append(*reqBuffer.buffer, body)
		reqBuffer.mu.Unlock()
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
