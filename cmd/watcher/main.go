package main

import (
	"log"
	"os/exec"
	"strings"
	"time"
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
		log.Println("MEM:")
		var memSlice = getMemorySlice(string(free))
		var memTotal = memSlice[0]
		var memUsed = memSlice[1]
		var memFree = memSlice[2]
		log.Printf("%s / %s ( %s )", memUsed, memTotal, memFree)
		log.Println("CPU:")
		var cpuSlice = getCpuSlice(string(cpu))
		for _, core := range cpuSlice {
			var user = core[0]
			var niced = core[1]
			var system = core[2]
			var idle = core[3]
			log.Printf("%s %s %s / %s", user, niced, system, idle)
		}

		// TODO: pack structs
		// send requests to the client service

		time.Sleep(time.Duration(3) * time.Second)
	}
}
