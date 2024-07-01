package util

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"
)

type ClientRequestBuffer struct {
	Mu     sync.Mutex
	Buffer *[]ClientMessage
}

type RouterRequestBuffer struct {
	Mu     sync.Mutex
	Buffer *[]string
}

func ClientSendLoop(reqBuffer *ClientRequestBuffer, address string) {
	var baseSleep = 1000
	var sleepFactor = 1
	client := &http.Client{}
	for {
		if len(*reqBuffer.Buffer) <= 0 {
			time.Sleep(1000 * time.Millisecond)
			continue
		}
		reqBuffer.Mu.Lock()
		req := (*reqBuffer.Buffer)[0]
		reqBuffer.Mu.Unlock()
		b, err := json.Marshal(req)
		if err != nil {
			log.Println("Unable to marshal ClientRequest.", err)
			continue
		}
		tmp, err := http.NewRequest("POST", address, bytes.NewBuffer(b))

		if err != nil {
			log.Println("Unable to create HTTP POST request", err)
			return
		}

		_, err = client.Do(tmp)

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
		reqBuffer.Mu.Lock()
		*reqBuffer.Buffer = (*reqBuffer.Buffer)[1:] // remove handled element from queue
		reqBuffer.Mu.Unlock()
	}
}
