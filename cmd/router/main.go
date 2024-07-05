package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/numyalai/fog-computing-assignment/pkg/util"
)

func deregisterInactiveClients(clientStorage *util.Storage) {
	for {
		clientStorage.DeregisterInactiveClients(1 * time.Minute)
		time.Sleep(1 * time.Minute)
	}
}

func main() {
	log.SetPrefix("router: ")
	log.Println("Starting ...")

	allocBuffer := make([]string, 0)
	var reqBuffer = util.RouterRequestBuffer{Buffer: &allocBuffer}
	server := http.NewServeMux()
	server.HandleFunc("/forward", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received %s request from %s", r.Method, r.RemoteAddr)
		buffer := new(bytes.Buffer)
		buffer.ReadFrom(r.Body)
		body := buffer.String()
		log.Println(body)
		reqBuffer.Mu.Lock()
		defer reqBuffer.Mu.Unlock()
		*reqBuffer.Buffer = append(*reqBuffer.Buffer, body)
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK\n"))
	})

	go func() {
		listenAddr := "0.0.0.0:6001"
		log.Printf("Serving HTTP at %s", listenAddr)
		err := http.ListenAndServe(listenAddr, server)
		if err != nil {
			log.Println("Unable to open HTTP server on %", listenAddr, err)
			os.Exit(1)
		}
	}()

	storage := util.NewStorage()
	go deregisterInactiveClients(storage)
	//go util.RouterSendLoop(storage, &reqBuffer)

	buf := make([]byte, 4096)
	addr := net.UDPAddr{
		Port: 5001,
		IP:   net.ParseIP("0.0.0.0"),
	}

	ser, err := net.ListenUDP("udp", &addr)
	if err != nil {
		log.Panic("Unable to open udp listening port. ", err)
		os.Exit(1)
	}
	log.Printf("Serving UDP endpoint at %s:%d", addr.IP, addr.Port)
	defer ser.Close()

	log.Println("Entering UDP main loop")
	for {
		n, raddr, err := ser.ReadFromUDP(buf)
		if err != nil {
			log.Println("Error in reading UDP occured. ", err)
			continue
		}
		req := util.SentPackage{}

		err = json.Unmarshal(buf[:n], &req)
		if err != nil {
			log.Println("Unable to umarshal received UDP request. ", err)
		}

		t := util.ClientMessage{}
		err = json.Unmarshal(req.Data, &t)
		if err != nil {
			log.Println("Unable to unmarshal HTTP request from client.", err)
		}
		log.Println(t)
		storage.RegisterClient(raddr, t.Data.Memory, t.Data.Cpu)
		storage.UpdateClient(raddr, t.Data.Memory, t.Data.Cpu)

		tmp := util.AckPackage{
			Id: req.Id,
		}
		log.Println(string(buf[:n]))

		resp, err := json.Marshal(tmp)
		if err != nil {
			log.Println("Unable to Marshal ack package. ", err)
		}
		_, err = ser.WriteToUDP(resp, raddr)

		if err != nil {
			log.Println(err)
		}
	}

	/*storage := util.NewStorage()

	buf := make([]string, 0)
	var reqBuffer = util.RouterRequestBuffer{Buffer: &buf}
	go util.RouterSendLoop(storage, &reqBuffer)
	go deregisterInactiveClients(storage)

	server.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received %s request from %s", r.Method, r.RemoteAddr)
		buffer := &bytes.Buffer{}
		buffer.ReadFrom(r.Body)
		t := util.ClientMessage{}
		err := json.Unmarshal(buffer.Bytes(), &t)
		if err != nil {
			log.Println("Unable to unmarshal HTTP request from client.", err)
		}
		log.Println(t)
		cliendID := t.Endpoint
		storage.RegisterClient(cliendID, t.Data.Memory, t.Data.Cpu)
		storage.UpdateClient(cliendID, t.Data.Memory, t.Data.Cpu)

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
		defer reqBuffer.Mu.Unlock()
		*reqBuffer.Buffer = append(*reqBuffer.Buffer, body)
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK\n"))
	})

	log.Printf("Serving at %s", listenAddr)
	err := http.ListenAndServe(listenAddr, server)

	if err != nil {
		log.Printf("%s", err)
	}*/

	log.Println("Stopped.")
}
