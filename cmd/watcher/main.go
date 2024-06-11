package main

import (
	"log"
	"os/exec"
	"time"
)

func main() {
	log.SetPrefix("router: ")
	log.Println("Starting ...")

	var clientEndpoint = "http://localhost:5002"
	log.Println(clientEndpoint)
	for {
		cmd := exec.Command("free")
		free, err := cmd.Output()

		if err != nil {
			log.Println(err)
		}

		cmd = exec.Command("cat", "/proc/stat")
		cpu, err := cmd.Output()

		if err != nil {
			log.Println(err)
		}

		log.Println(string(free))
		log.Println(string(cpu))

		// TODO: parse free and cpu
		// pack them into structs
		// send requests to the client service

		time.Sleep(time.Duration(3) * time.Second)
	}
}
