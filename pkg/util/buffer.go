package util

import (
	"bytes"
	"log"
	"net/http"
	"sync"
	"time"
)

type RequestBuffer struct {
	Mu     sync.Mutex
	Buffer *[]string
}

func SendLoop(reqBuffer *RequestBuffer, address string) {
	var baseSleep = 1000
	var sleepFactor = 1
	client := &http.Client{}
	for {
		if len(*reqBuffer.Buffer) <= 0 {
			var sleepDuration = baseSleep * sleepFactor
			time.Sleep(time.Duration(sleepDuration) * time.Millisecond)
			if sleepFactor < 120 {
				sleepFactor = sleepFactor * 2
			}
			continue
		}
		reqBuffer.Mu.Lock()
		req := (*reqBuffer.Buffer)[0]
		log.Println("Request := ", req)
		tmp, err := http.NewRequest("POST", address, bytes.NewBufferString(req))

		if err != nil {
			reqBuffer.Mu.Unlock()
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
			reqBuffer.Mu.Unlock()
			continue
		}

		log.Println(resp)

		*reqBuffer.Buffer = (*reqBuffer.Buffer)[1:] // remove handled element from queue
		reqBuffer.Mu.Unlock()
	}
}
