package util

import (
	"sync"
)

type RequestBuffer struct {
	mu     sync.Mutex
	buffer *[]string
}
