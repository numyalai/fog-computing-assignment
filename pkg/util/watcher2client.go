package util

type MemoryData struct {
	Free  uint64
	Total uint64
}

type CpuData struct {
	Free  uint64
	Total uint64
}

type Message struct {
	Memory MemoryData
	Cpu    CpuData
}
