package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
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

	allocBuffer := make([]util.PacketUDP, 0)
	var reqBuffer = util.RequestBuffer{Buffer: allocBuffer}
	server := http.NewServeMux()
	server.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		buffer := new(bytes.Buffer)
		buffer.ReadFrom(r.Body)
		body := buffer.String()
		log.Printf("Received %s request from %s with HTTP body: %s\n", r.Method, r.RemoteAddr, body)
		tmp := util.PacketUDP{
			Id:   uuid.New().String(),
			Data: []byte(body),
		}
		reqBuffer.Mu.Lock()
		reqBuffer.Buffer = append(reqBuffer.Buffer, tmp)
		reqBuffer.Mu.Unlock()
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

	acks := util.SafeAcks{
		Acks: make([]string, 0),
	}
	go util.RouterSendLoop(storage, &reqBuffer, ser, &acks)
	defer ser.Close()

	for {
		n, raddr, err := ser.ReadFromUDP(buf)
		if err != nil {
			log.Println("Error in reading UDP occured. ", err)
			continue
		}
		req := util.PacketUDP{}

		err = json.Unmarshal(buf[:n], &req)
		if err != nil {
			log.Println("Unable to umarshal received UDP request. ", err)
		}

		if req.Data != nil {
			t := util.WatcherMessage{}
			err = json.Unmarshal(req.Data, &t)
			if err != nil {
				log.Println("Unable to unmarshal HTTP request from client.", err)
			}
			storage.RegisterOrUpdateClient(raddr, t.Memory, t.Cpu)

			tmp := util.PacketUDP{
				Id:   req.Id,
				Data: nil,
			}

			resp, err := json.Marshal(tmp)
			if err != nil {
				log.Println("Unable to Marshal ack package. ", err)
			}
			_, err = ser.WriteToUDP(resp, raddr)

			if err != nil {
				log.Println(err)
			}
		} else {
			acks.Mu.Lock()
			acks.Acks = append(acks.Acks, req.Id)
			acks.Mu.Unlock()
		}
	}

	log.Println("Stopped.")
}
