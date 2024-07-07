package util

type ClientMessage struct {
	Endpoint string
	Data     WatcherMessage
}
type MemoryData struct {
	Free  uint64
	Total uint64
}

type CpuData struct {
	Free  uint64
	Total uint64
}

type WatcherMessage struct {
	Memory MemoryData
	Cpu    CpuData
}
