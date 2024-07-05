package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/numyalai/fog-computing-assignment/pkg/util"
)

func PrintUsage() {
	log.Printf("Usage: %s <IP_OF_CLIENT> <IP_OF_ROUTER>", os.Args[0])
}

func main() {
	log.SetPrefix("client: ")
	log.Println("Starting ...")

	argsWithoutProg := os.Args[1:]

	listenAddr := fmt.Sprintf("%s:5002", argsWithoutProg[0])
	if listenAddr == ":5002" {
		PrintUsage()
		os.Exit(1)
	}
	if argsWithoutProg[1] == "" {
		PrintUsage()
		os.Exit(2)
	}
	routerEndpoint := argsWithoutProg[1]

	server := http.NewServeMux()

	buf := make([]util.ClientMessage, 0)
	var reqBuffer = util.ClientRequestBuffer{Buffer: &buf}
	go util.ClientSendLoop(&reqBuffer, routerEndpoint)

	server.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received %s request from %s", r.Method, r.RemoteAddr)
		buffer := &bytes.Buffer{}
		buffer.ReadFrom(r.Body)
		t := util.WatcherMessage{}
		err := json.Unmarshal(buffer.Bytes(), &t)
		if err != nil {
			log.Println("Unable to parse requests body: ", err)
		}
		msg := util.ClientMessage{
			Endpoint: listenAddr + "/forward",
			Data:     t,
		}
		log.Println(t)
		reqBuffer.Mu.Lock()
		*reqBuffer.Buffer = append(*reqBuffer.Buffer, msg)
		reqBuffer.Mu.Unlock()
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK\n"))
	})

	server.HandleFunc("/forward", func(w http.ResponseWriter, r *http.Request) {
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

	log.Printf("Serving at %s", listenAddr)
	err := http.ListenAndServe(listenAddr, server)

	if err != nil {
		log.Printf("%s", err)
	}

	log.Println("Stopped.")
}
