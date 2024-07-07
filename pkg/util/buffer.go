package util

import (
	"sync"
)

type RequestBuffer struct {
	Mu     sync.Mutex
	Buffer []PacketUDP
}
