package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/numyalai/fog-computing-assignment/pkg/util"
)

func PrintUsage() {
	log.Printf("Usage: %s <IP_OF_ROUTER>:<PORT_OF_ROUTER>", os.Args[0])
}

func main() {
	log.SetPrefix("client: ")
	log.Println("Starting ...")

	if len(os.Args) < 2 {
		PrintUsage()
		os.Exit(1)
	}

	argsWithoutProg := os.Args[1:]

	listenAddr := "0.0.0.0:5002"
	if argsWithoutProg[0] == "" {
		PrintUsage()
		os.Exit(2)
	}
	routerEndpoint := argsWithoutProg[0]
	if !strings.Contains(routerEndpoint, ":") {
		log.Printf("Input: '%s'", routerEndpoint)
		PrintUsage()
		os.Exit(6)
	}

	server := http.NewServeMux()

	buf := make([]util.PacketUDP, 0)
	var reqBuffer = util.RequestBuffer{Buffer: buf}
	var packets = util.SafeBuffer{
		Data: make([]util.PacketUDP, 0),
	}
	go util.RouterConnection(&reqBuffer, routerEndpoint, &packets)

	server.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Forwarding message -> router@%s", routerEndpoint)
		buffer := &bytes.Buffer{}
		buffer.ReadFrom(r.Body)
		msg := util.PacketUDP{
			Id:   uuid.New().String(),
			Data: buffer.Bytes(),
		}
		reqBuffer.Mu.Lock()
		reqBuffer.Buffer = append(reqBuffer.Buffer, msg)
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

	go func() {
		log.Printf("Serving at %s", listenAddr)
		err := http.ListenAndServe(listenAddr, server)

		if err != nil {
			log.Printf("%s", err)
		}
	}()

	for {
		if len(packets.Data) <= 0 {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		packets.Mu.Lock()
		packet := packets.Data[0]
		packets.Data = packets.Data[1:]
		packets.Mu.Unlock()
		stressAddress := "localhost:5010"
		client := http.Client{Timeout: 5000 * time.Millisecond}
		body := bytes.NewBuffer(packet.Data)
		for {
			log.Println(packet.Id)
			resp, err := client.Post(fmt.Sprintf("http://%s", stressAddress), "plain/text", body)
			if err != nil {
				log.Println("Cannot go stress the system :'(", err)
				time.Sleep(1 * time.Second)
				continue
			}
			if resp.StatusCode >= 400 {
				log.Println("Error in processing forwarded stress message.")
				time.Sleep(100 * time.Millisecond)
				continue
			}
			log.Printf("Forwarding %s -> %s", string(packet.Data), stressAddress)
			break
		}
	}
}
