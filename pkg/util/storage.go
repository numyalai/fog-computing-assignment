package util

import (
	"fmt"
	"net"
	"sync"
	"time"
)

type Client struct {
	RAM       MemoryData
	CPU       CpuData
	Address   *net.UDPAddr
	UpdatedAt time.Time
}

type Storage struct {
	Mu      sync.Mutex
	Storage map[string]*Client
}

func NewStorage() *Storage {
	return &Storage{
		Storage: make(map[string]*Client),
	}
}

func getUDPAddrRepresentation(addr *net.UDPAddr) string {
	return fmt.Sprintf("%s:%d", addr.IP.String(), addr.Port)
}

func (cs *Storage) RegisterClient(clientID *net.UDPAddr, ram MemoryData, cpu CpuData) {
	cs.Mu.Lock()
	defer cs.Mu.Unlock()
	cs.Storage[getUDPAddrRepresentation(clientID)] = &Client{
		RAM:       ram,
		CPU:       cpu,
		Address:   clientID,
		UpdatedAt: time.Now(),
	}
}

func (s *Storage) UpdateClient(id *net.UDPAddr, ram MemoryData, cpu CpuData) {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	if client, exists := s.Storage[getUDPAddrRepresentation(id)]; exists {
		client.RAM = ram
		client.CPU = cpu
		client.UpdatedAt = time.Now()
	}
}

func (s *Storage) GetClient(id *net.UDPAddr) *Client {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	if client, exists := s.Storage[getUDPAddrRepresentation(id)]; exists {
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

func (cs *Storage) GetAllClients() map[string]*Client {
	cs.Mu.Lock()
	defer cs.Mu.Unlock()
	return cs.Storage
}
