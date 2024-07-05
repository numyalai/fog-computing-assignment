package util

import (
	"bytes"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
)

type ClientRequestBuffer struct {
	Mu     sync.Mutex
	Buffer *[]ClientMessage
}

type RouterRequestBuffer struct {
	Mu     sync.Mutex
	Buffer *[]string
}

type SentPackage struct {
	Id   string
	Data []byte
}

type AckPackage struct {
	Id string
}

func Send2Client(conn *net.UDPConn, data string) error {
	return Send(conn, []byte(data))
}

func Send2Router(conn *net.UDPConn, data ClientMessage) error {
	b, err := json.Marshal(data)
	if err != nil {
		log.Println("Unable to marshal. ", err)
		time.Sleep(1 * time.Second)
		Send2Router(conn, data)
	}
	return Send(conn, b)
}

func Send(conn *net.UDPConn, data []byte) error {
	body := SentPackage{
		Id:   uuid.New().String(),
		Data: data,
	}
	udpBody, err := json.Marshal(body)
	if err != nil {
		log.Println("Unable to marshal")
		return err
	}
	err = SendMessage(conn, udpBody)
	for err != nil {
		log.Println("Retrying after 1 second 1")
		time.Sleep(1 * time.Second)
		err = SendMessage(conn, udpBody)
	}
	messageID, readErr := receiveAck(conn)
	if readErr != nil {
		log.Println("Unable to read ack. ", readErr)
		return readErr
	}
	log.Printf("%s == %s\n", messageID, body.Id)
	if messageID == body.Id {
		return nil
	}
	log.Println("Retrying to send data, not acknowledged.")
	return Send(conn, data)
}

func receiveAck(conn *net.UDPConn) (string, error) {
	buf := make([]byte, 4096)
	ack := AckPackage{}
	n, _, err := conn.ReadFromUDP(buf)
	if err != nil {
		log.Println("Unable to read from UDP connection.", err)
		return "", err
	}
	err = json.Unmarshal(buf[:n], &ack)
	if err != nil {
		log.Println("Unable to unmarshal ack apckage. ", err)
		return "", err
	}
	return ack.Id, nil
}

func SendMessage(conn *net.UDPConn, msg []byte) error {
	var tryNumber = 0
	for {
		_, err := conn.Write(msg)
		if err != nil {
			if tryNumber > 10 {
				return err
			}
			tryNumber += 1
			log.Println("Continuing after 1 second")
			time.Sleep(100 * time.Millisecond)
			continue
		}
		return nil
	}
}

func ClientSendLoop(reqBuffer *ClientRequestBuffer, address string) {
	var baseSleep = 1000
	var sleepFactor = 1

	addr := net.UDPAddr{
		Port: 5001,
		IP:   net.ParseIP(address),
	}

	for {
		c, err := net.DialUDP("udp", nil, &addr)
		if err != nil {
			log.Panic("Unable to establish udp connection. ", err)
			os.Exit(3)
		}
		if len(*reqBuffer.Buffer) <= 0 {
			time.Sleep(1000 * time.Millisecond)
			continue
		}
		reqBuffer.Mu.Lock()
		req := (*reqBuffer.Buffer)[0]
		reqBuffer.Mu.Unlock()

		err = Send2Router(c, req)

		if err != nil {
			var sleepDuration = baseSleep * sleepFactor
			log.Printf("Retrying after %d seconds", sleepDuration/1000)
			time.Sleep(time.Duration(sleepDuration) * time.Millisecond)
			if sleepFactor < 120 {
				sleepFactor = sleepFactor * 2
			}
			continue
		}

		sleepFactor = 1
		reqBuffer.Mu.Lock()
		*reqBuffer.Buffer = (*reqBuffer.Buffer)[1:] // remove handled element from queue
		reqBuffer.Mu.Unlock()
	}
}

func RouterSendLoop(storage *Storage, reqBuffer *RouterRequestBuffer) {
	var baseSleep = 1000
	var sleepFactor = 1
	client := &http.Client{}
	for {
		if len(*reqBuffer.Buffer) <= 0 {
			time.Sleep(1000 * time.Millisecond)
			continue
		}
		var address string
		var cpu float64
		var ram float64
		for endpoint, data := range storage.GetAllClients() {
			var cpuAvailability float64 = float64(data.CPU.Free) / float64(data.CPU.Total) * 100.0
			var ramAvailability float64 = float64(data.RAM.Free) / float64(data.RAM.Total) * 100.0
			log.Printf("ENTRY := %s CPU: %f%% RAM: %f%%", endpoint, cpuAvailability, ramAvailability)
			if cpu > 10.0 && ram > 10.0 && (cpuAvailability > cpu && ramAvailability > ram || cpuAvailability > cpu && ramAvailability > 20.0) {
				cpu = cpuAvailability
				ram = ramAvailability
				address = endpoint
			}
		}
		if address != "" {
			log.Println("Unable to find suitable edge node. Retrying in 1 second.")
			time.Sleep(1000 * time.Millisecond)
			continue
		}
		reqBuffer.Mu.Lock()
		req := (*reqBuffer.Buffer)[0]
		reqBuffer.Mu.Unlock()
		tmp, err := http.NewRequest("POST", address, bytes.NewBufferString(req))

		if err != nil {
			log.Println("Unable to create HTTP POST request", err)
			return
		}

		resp, err := client.Do(tmp)

		if err != nil {
			var sleepDuration = baseSleep * sleepFactor
			log.Printf("Failed to perform HTTP request. Retrying after %d seconds.", sleepDuration/1000)
			time.Sleep(time.Duration(sleepDuration) * time.Millisecond)
			if sleepFactor < 120 {
				sleepFactor = sleepFactor * 2
			}
			continue
		}

		log.Println(resp)
		sleepFactor = 1
		reqBuffer.Mu.Lock()
		*reqBuffer.Buffer = (*reqBuffer.Buffer)[1:] // remove handled element from queue
		reqBuffer.Mu.Unlock()
	}
}
