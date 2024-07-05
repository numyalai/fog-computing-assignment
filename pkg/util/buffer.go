package util

import (
	"encoding/json"
	"log"
	"net"
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

type PacketUDP struct {
	Id   string
	Data []byte
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
	body := PacketUDP{
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
	if messageID == body.Id {
		return nil
	}
	log.Println("Retrying to send data, not acknowledged.")
	return Send(conn, data)
}

func receiveAck(conn *net.UDPConn) (string, error) {
	buf := make([]byte, 4096)
	ack := PacketUDP{}
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
		c, err := net.DialUDP("udp", nil, &addr) // TODO: rework listening on this and discriminate between ACKS and forwards
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

func RouterSendLoop(storage *Storage, reqBuffer *RouterRequestBuffer, socket *net.UDPConn) {
	var baseSleep = 1000
	var sleepFactor = 1
	for {
		if len(*reqBuffer.Buffer) <= 0 {
			time.Sleep(1000 * time.Millisecond)
			continue
		}
		var address *net.UDPAddr
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
		if address != nil {
			log.Println("Unable to find suitable edge node. Retrying in 1 second.")
			time.Sleep(1000 * time.Millisecond)
			continue
		}
		reqBuffer.Mu.Lock()
		req := (*reqBuffer.Buffer)[0]
		reqBuffer.Mu.Unlock()
		_, err := socket.WriteToUDP([]byte(req), address) // TODO: loop and retry after some fails with new routing decision

		if err != nil {
			var sleepDuration = baseSleep * sleepFactor
			log.Printf("Failed to forward datagram. Retrying after %d seconds.", sleepDuration/1000)
			time.Sleep(time.Duration(sleepDuration) * time.Millisecond)
			if sleepFactor < 120 {
				sleepFactor = sleepFactor * 2
			}
			continue
		}

		sleepFactor = 1
		reqBuffer.Mu.Lock()
		for i, r := range *reqBuffer.Buffer {
			if r == req {
				(*reqBuffer.Buffer) = append((*reqBuffer.Buffer)[:i], (*reqBuffer.Buffer)[:i+1]...)
				break
			}
		}
		reqBuffer.Mu.Unlock()
	}
}
