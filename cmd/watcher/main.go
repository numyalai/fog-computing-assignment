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

type CpuCore struct {
	User   uint64
	Niced  uint64
	System uint64
	Idle   uint64
}

type Cpu struct {
	Cores []CpuCore
}

type DeltaCpu struct {
	Cores []DeltaCpuCore
}

type DeltaCpuCore struct {
	Idle      uint64
	Execution uint64
}

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

func getCpuSlice(cpuOutput string) Cpu {
	var cpuStat = Cpu{
		Cores: make([]CpuCore, 0),
	}
	var cpuLines = strings.Split(cpuOutput, "\n")
	for _, line := range cpuLines {
		var skip = false
		var index = -1
		var core = CpuCore{}
		if !strings.Contains(line, "cpu") {
			break
		}
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
		tmpUser, err := strconv.ParseUint(fields[0], 10, 64)
		if err != nil {
			log.Panicln("Unable to parse user cycles of CPU", err)
			continue
		}
		tmpNiced, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			log.Panicln("Unable to parse user niced cycles of CPU", err)
			continue
		}
		tmpSystem, err := strconv.ParseUint(fields[2], 10, 64)
		if err != nil {
			log.Panicln("Unable to parse system cycles of CPU", err)
			continue
		}
		tmpIdle, err := strconv.ParseUint(fields[3], 10, 64)
		if err != nil {
			log.Panicln("Unable to parse idle cycles of CPU", err)
			continue
		}

		core.User = tmpUser
		core.Niced = tmpNiced
		core.System = tmpSystem
		core.Idle = tmpIdle
		cpuStat.Cores = append(cpuStat.Cores, core)
	}
	return cpuStat
}

func getDeltaCpu(previous, current Cpu) DeltaCpu {
	var cpu = DeltaCpu{
		Cores: make([]DeltaCpuCore, 0),
	}

	for i := 0; i < len(previous.Cores); i++ {
		var pCore CpuCore = previous.Cores[i]
		var cCore CpuCore = current.Cores[i]
		previousExecutionCycles := pCore.User + pCore.Niced + pCore.System
		executionCycles := cCore.User + cCore.Niced + cCore.System

		previousIdle := pCore.Idle
		currentIdle := cCore.Idle

		previousSum := previousIdle + previousExecutionCycles
		currentSum := currentIdle + executionCycles

		deltaTotal := currentSum - previousSum
		deltaIdle := currentIdle - previousIdle
		cpu.Cores = append(cpu.Cores, DeltaCpuCore{
			Execution: deltaTotal,
			Idle:      deltaIdle,
		})
	}
	return cpu
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

		cmd = exec.Command("cat", "/proc/stat")
		cpuProcStat, err := cmd.Output()

		if err != nil {
			log.Println(err)
		}

		var previous = getCpuSlice(string(cpuProcStat))
		time.Sleep(1 * time.Second)
		cmd = exec.Command("cat", "/proc/stat")
		cpuProcStat, err = cmd.Output()

		if err != nil {
			log.Println(err)
		}
		var current = getCpuSlice(string(cpuProcStat))

		var dCpu = getDeltaCpu(previous, current)

		var idle, total uint64

		for _, dCore := range dCpu.Cores {
			idle += dCore.Idle
			total += dCore.Execution
		}

		var cpuData = util.CpuData{
			Free:  idle,
			Total: total,
		}
		var tmp = util.WatcherMessage{
			Memory: memData,
			Cpu:    cpuData,
		}
		body, err := json.Marshal(tmp)
		if err != nil {
			log.Println("Unable to marshal body message.", err)
		}
		req, err := http.NewRequest("POST", clientEndpoint, bytes.NewBuffer(body))
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
