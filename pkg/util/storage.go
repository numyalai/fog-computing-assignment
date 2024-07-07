package util

import (
	"fmt"
	"log"
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

func (s *Storage) RegisterOrUpdateClient(id *net.UDPAddr, ram MemoryData, cpu CpuData) {
	s.Mu.Lock()

	if client, exists := s.Storage[getUDPAddrRepresentation(id)]; exists {
		if ram.Total != 0 {
			log.Printf("Updating RAM for %s", id)
			client.RAM = ram
		}
		if cpu.Total != 0 {
			log.Printf("Updating CPU for %s", id)
			client.CPU = cpu
		}
		client.UpdatedAt = time.Now()
		s.Mu.Unlock()
	} else {
		s.Mu.Unlock()
		s.RegisterClient(id, ram, cpu)
	}
}

func (cs *Storage) RegisterClient(id *net.UDPAddr, ram MemoryData, cpu CpuData) {
	cs.Mu.Lock()
	defer cs.Mu.Unlock()
	client := Client{
		Address:   id,
		UpdatedAt: time.Now(),
	}
	if ram.Total != 0 {
		log.Printf("Storing CPU for %s", id)
		client.RAM = ram
	}
	if cpu.Total != 0 {
		log.Printf("Storing CPU for %s", id)
		client.CPU = cpu
	}
	cs.Storage[getUDPAddrRepresentation(id)] = &client
}

func (s *Storage) UpdateClient(id *net.UDPAddr, ram MemoryData, cpu CpuData) {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	if client, exists := s.Storage[getUDPAddrRepresentation(id)]; exists {
		if ram.Total != 0 {
			client.RAM = ram
		}
		if cpu.Total != 0 {
			client.CPU = cpu
		}
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
