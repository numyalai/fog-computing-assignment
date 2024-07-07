package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/numyalai/fog-computing-assignment/pkg/util"
)

func getMemorySlice(freeOutput string) []string {
	var freeCmdOutput = strings.Split(freeOutput, "\n")[1]
	var memStat = make([]string, 0)
	var tmp = strings.TrimSpace(strings.TrimLeft(freeCmdOutput, "Mem:"))
	var index = 0
	for i, char := range tmp {
		if index == -1 && char != ' ' {
			index = i
		}
		if index > -1 && char == ' ' {
			memStat = append(memStat, tmp[index:i])
			index = -1
		}
	}
	return memStat
}

func main() {
	log.SetPrefix("ram_watcher: ")
	log.Println("Starting ...")

	var clientEndpoint = "http://localhost:5002"

	log.Printf("Directing messages at %s\n", clientEndpoint)
	client := &http.Client{}
	for {
		cmd := exec.Command("free")
		free, err := cmd.Output()

		if err != nil {
			log.Println(err)
		}

		var memSlice = getMemorySlice(string(free))
		memTotal, err := strconv.ParseUint(memSlice[0], 10, 64)
		if err != nil {
			log.Panicln("Unable to parse total available memory.", err)
			time.Sleep(time.Duration(3) * time.Second)
			continue
		}
		memFree, err := strconv.ParseUint(memSlice[2], 10, 64)
		if err != nil {
			log.Panicln("Unable to parse free available memory", err)
			continue
		}
		var memData = util.MemoryData{
			Free:  memFree,
			Total: memTotal,
		}

		var tmp = util.WatcherMessage{
			Memory: memData,
		}
		body, err := json.Marshal(tmp)
		if err != nil {
			log.Println("Unable to marshal body message.", err)
			continue
		}
		req, err := http.NewRequest("POST", clientEndpoint, bytes.NewBuffer(body))
		if err != nil {
			log.Println("Unable to creat HTTP POST request", err)
			continue
		}
		resp, err := client.Do(req)
		if err != nil {
			log.Println("Unable to perform HTTP POST request to client", err)
			continue
		}
		if resp.StatusCode >= 400 {
			log.Printf("Error in request handeling at client. HTTP %d %s", resp.StatusCode, resp.Status)
			continue
		}
		log.Printf("MEM status: Available %.2f%%", float64(memData.Free)/float64(memData.Total)*100.0)
		time.Sleep(time.Duration(1) * time.Second) // might need to change the timescaling here later
	}
}
