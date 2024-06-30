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

func getCpuSlice(cpuOutput string) [][]string {
	var cpuStat = make([][]string, 0)
	var cpuLines = strings.Split(cpuOutput, "\n")
	for i, line := range cpuLines {
		var skip = false
		var index = -1
		for pos, char := range line {
			if !skip && char != ' ' {
				continue
			}
			if char == ' ' {
				skip = true
				continue
			}
			if char != ' ' {
				index = pos
				break
			}
		}
		var fields = strings.Split(line[index:], " ")
		cpuStat = append(cpuStat, fields[:4])
		if i >= 8 {
			break
		}
	}
	return cpuStat
}

func main() {
	log.SetPrefix("watcher: ")
	log.Println("Starting ...")

	var clientEndpoint = "http://localhost:5002"

	client := &http.Client{}
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
		var cpuSlice = getCpuSlice(string(cpu))
		var core = cpuSlice[0]
		user, err := strconv.ParseUint(core[0], 10, 64)
		if err != nil {
			log.Panicln("Unable to parse user cycles of CPU", err)
			continue
		}
		niced, err := strconv.ParseUint(core[1], 10, 64)
		if err != nil {
			log.Panicln("Unable to parse user niced cycles of CPU", err)
			continue
		}
		system, err := strconv.ParseUint(core[2], 10, 64)
		if err != nil {
			log.Panicln("Unable to parse system cycles of CPU", err)
			continue
		}
		idle, err := strconv.ParseUint(core[3], 10, 64)
		if err != nil {
			log.Panicln("Unable to parse idle cycles of CPU", err)
			continue
		}

		var cpuData = util.CpuData{
			Free:  idle,
			Total: user + niced + system + idle,
		}
		var tmp = util.Message{
			Memory: memData,
			Cpu:    cpuData,
		}
		body := new(bytes.Buffer)
		json.NewEncoder(body).Encode(tmp)
		req, err := http.NewRequest("POST", clientEndpoint, body)
		if err != nil {
			log.Panicln("Unable to creat HTTP POST request", err)
			continue
		}
		resp, err := client.Do(req)
		if err != nil {
			log.Panicln("Unable to perform HTTP POST request to client", err)
			continue
		}
		if resp.StatusCode >= 400 {
			log.Panicf("Error in request handeling at client. HTTP %d %s", resp.StatusCode, resp.Status)
			continue
		}
		log.Printf("Sent sytem status to %s", clientEndpoint)
		time.Sleep(time.Duration(3) * time.Second) // might need to change the timescaling here later
	}
}
