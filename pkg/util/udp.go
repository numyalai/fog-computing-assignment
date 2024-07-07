package util

import (
	"encoding/json"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type PacketUDP struct {
	Id   string
	Data []byte
}

type SafeAcks struct {
	Mu   sync.Mutex
	Acks []string
}

type SafeBuffer struct {
	Mu   sync.Mutex
	Data []PacketUDP
}

func Send2Router(conn *net.UDPConn, data PacketUDP, acks *SafeAcks) error {
	b, err := json.Marshal(data)
	if err != nil {
		log.Println("Unable to marshal. ", err)
		return err
	}
	return Send(conn, b, data.Id, acks)
}

func Send(conn *net.UDPConn, data []byte, id string, acks *SafeAcks) error {
	for i := 0; i < 10; i++ {
		err := SendMessage(conn, data)
		if err != nil {
			log.Println("Retrying... ", err)
			time.Sleep(100 * time.Millisecond)
			continue
		}
		for i := 0; i < 10; i++ {
			time.Sleep(100 * time.Millisecond)
			acks.Mu.Lock()
			data := acks.Acks
			acks.Mu.Unlock()
			for _, ack := range data {
				if ack == id {
					return nil
				}
			}
		}
	}
	return &net.OpError{}
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

func RouterConnection(reqBuffer *RequestBuffer, address string, packets *SafeBuffer) {
	var baseSleep = 1000
	var sleepFactor = 1

	sections := strings.Split(address, ":")
	port, err := strconv.Atoi(sections[1])

	if err != nil {
		log.Printf("No valid ports was passed. '%s' is not a valid port.", sections[1])
		os.Exit(5)
	}

	addr := net.UDPAddr{
		Port: port,
		IP:   net.ParseIP(sections[0]),
	}

	acks := SafeAcks{
		Acks: make([]string, 0),
	}

	for {
		c, err := net.DialUDP("udp", nil, &addr)
		if err != nil {
			log.Println("ERROR: Unable to establish udp connection. ", err)
			var sleepDuration = baseSleep * sleepFactor
			log.Printf("Retrying after %d seconds", sleepDuration/1000)
			time.Sleep(time.Duration(sleepDuration) * time.Millisecond)
			if sleepFactor <= 16 {
				sleepFactor = sleepFactor * 2
			}
			continue
		}
		sleepFactor = 1
		var running = true
		go handleRouterResponses(c, &acks, packets, &running)
		handleBufferPacketSend(c, reqBuffer, &acks)
		running = false
		time.Sleep(100 * time.Millisecond)
	}
}

func handleRouterResponses(socket *net.UDPConn, acks *SafeAcks, packets *SafeBuffer, running *bool) {
	buf := make([]byte, 65535)
	for *running {
		input := PacketUDP{}
		n, _, err := socket.ReadFromUDP(buf)
		if err != nil {
			log.Println("Unable to read from UDP socket. ", err)
			continue
		}
		err = json.Unmarshal(buf[:n], &input)
		if err != nil {
			log.Println("Unable to unmarshal buffer. ", err)
			continue
		}
		if input.Data == nil {
			acks.Mu.Lock()
			acks.Acks = append(acks.Acks, input.Id)
			acks.Mu.Unlock()
		} else {
			tmp := PacketUDP{
				Id: input.Id,
			}
			body, err := json.Marshal(tmp)
			if err != nil {
				log.Println("Unable to marshal ACK packet. ", err)
			}
			_, err = socket.Write(body)
			if err != nil {
				log.Println("Unable to send ACK packet. ", err)
				continue
			}
			packets.Mu.Lock()
			packets.Data = append(packets.Data, input)
			packets.Mu.Unlock()
		}
	}
}

func handleBufferPacketSend(socket *net.UDPConn, reqBuffer *RequestBuffer, acks *SafeAcks) {
	var sleepFactor = 1
	var baseSleep = 1000
	for {
		if len(reqBuffer.Buffer) <= 0 {
			time.Sleep(1000 * time.Millisecond)
			continue
		}
		reqBuffer.Mu.Lock()
		req := (reqBuffer.Buffer)[0]
		reqBuffer.Mu.Unlock()

		err := Send2Router(socket, req, acks)

		if err != nil {
			var sleepDuration = baseSleep * sleepFactor
			log.Printf("Retrying after %d seconds", sleepDuration/1000)
			time.Sleep(time.Duration(sleepDuration) * time.Millisecond)
			if sleepFactor <= 16 {
				sleepFactor = sleepFactor * 2
				continue
			}
			break
		}

		sleepFactor = 1
		reqBuffer.Mu.Lock()
		reqBuffer.Buffer = removePacketUDP(reqBuffer.Buffer, req)
		reqBuffer.Mu.Unlock()
	}
}

func RouterSendBufferHandler(req PacketUDP, socket *net.UDPConn, address net.UDPAddr, acks *SafeAcks) error {
	var baseSleep = 1000
	var sleepFactor = 1

	for {
		body, err := json.Marshal(req)

		if err != nil {
			log.Println("Unable to marshal UDP packet.")
			time.Sleep(1000)
			continue
		}

		_, err = socket.WriteToUDP(body, &address)

		if err != nil {
			var sleepDuration = baseSleep * sleepFactor
			log.Println(err)
			log.Printf("Failed to forward datagram. Retrying after %d seconds.", sleepDuration/1000)
			time.Sleep(time.Duration(sleepDuration) * time.Millisecond)
			if sleepFactor <= 16 {
				sleepFactor = sleepFactor * 2
				continue
			}
			break
		}

		for i := 0; i < 10; i++ {
			time.Sleep(100 * time.Millisecond)
			acks.Mu.Lock()
			data := acks.Acks
			acks.Mu.Unlock()
			for _, ack := range data {
				if ack == req.Id {
					acks.Mu.Lock()
					acks.Acks = removeString(acks.Acks, req.Id)
					acks.Mu.Unlock()
					return nil
				}
			}
		}
		break
	}
	return net.ErrWriteToConnected
}

func RouterSendLoop(storage *Storage, reqBuffer *RequestBuffer, socket *net.UDPConn, acks *SafeAcks) {
	for {
		if len(reqBuffer.Buffer) <= 0 {
			time.Sleep(1000 * time.Millisecond)
			continue
		}
		var address net.UDPAddr
		var cpu float64 = 0.0
		var ram float64 = 0.0
		var found = false
		for _, data := range storage.GetAllClients() {
			if data.CPU.Total == 0 || data.RAM.Total == 0 {
				continue
			}
			var cpuAvailability float64 = float64(data.CPU.Free) / float64(data.CPU.Total) * 100.0
			var ramAvailability float64 = float64(data.RAM.Free) / float64(data.RAM.Total) * 100.0
			if cpuAvailability > 10.0 && ramAvailability > 10.0 && (cpuAvailability > cpu && ramAvailability > ram || cpuAvailability > cpu && ramAvailability > 20.0) {
				cpu = cpuAvailability
				ram = ramAvailability
				address = *data.Address
				found = true
			}
		}
		if !found {
			log.Println("Unable to find suitable edge node. Retrying in 1 second.")
			time.Sleep(1000 * time.Millisecond)
			continue
		}
		log.Printf("Forwarding request -> %s", address.String())

		reqBuffer.Mu.Lock()
		req := (reqBuffer.Buffer)[0]
		reqBuffer.Mu.Unlock()

		err := RouterSendBufferHandler(req, socket, address, acks)

		if err != nil {
			log.Println("Unable to send Buffer packet.")
			continue
		}

		reqBuffer.Mu.Lock()
		reqBuffer.Buffer = removePacketUDP(reqBuffer.Buffer, req)
		reqBuffer.Mu.Unlock()
		time.Sleep(500 * time.Millisecond)
	}
}
