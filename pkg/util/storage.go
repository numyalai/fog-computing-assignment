package util

import (
	"net"
	"sync"
	"time"
)

type Client struct {
	RAM        MemoryData
	CPU        CpuData
	Connection *net.UDPConn
	UpdatedAt  time.Time
}

type Storage struct {
	Mu      sync.Mutex
	Storage map[*net.UDPAddr]*Client
}

func NewStorage() *Storage {
	return &Storage{
		Storage: make(map[*net.UDPAddr]*Client),
	}
}

func (cs *Storage) RegisterClient(clientID *net.UDPAddr, ram MemoryData, cpu CpuData) {
	cs.Mu.Lock()
	defer cs.Mu.Unlock()
	cs.Storage[clientID] = &Client{
		RAM:       ram,
		CPU:       cpu,
		UpdatedAt: time.Now(),
	}
}

func (s *Storage) UpdateClient(id *net.UDPAddr, ram MemoryData, cpu CpuData) {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	if client, exists := s.Storage[id]; exists {
		client.RAM = ram
		client.CPU = cpu
		client.UpdatedAt = time.Now()
	}
}

func (s *Storage) GetClient(id *net.UDPAddr) *Client {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	if client, exists := s.Storage[id]; exists {
		return client
	}

	return nil
}

func (cs *Storage) DeregisterInactiveClients(timeout time.Duration) {
	// This is not needed as we are already locking the storage
	cs.Mu.Lock()
	defer cs.Mu.Unlock()
	for clientID, client := range cs.Storage {
		if time.Since(client.UpdatedAt) > timeout {
			delete(cs.Storage, clientID)
		}
	}
}

func (cs *Storage) GetAllClients() map[*net.UDPAddr]*Client {
	cs.Mu.Lock()
	defer cs.Mu.Unlock()
	return cs.Storage
}
