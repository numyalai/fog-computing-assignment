package util

import (
	"sync"
	"time"
)

type Client struct {
	ID        string
	RAM       string `json:"ram"`
	CPU       string `json:"cpu"`
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

func (cs *Storage) RegisterClient(clientID string, ram string, cpu string) {
	cs.Mu.Lock()
	defer cs.Mu.Unlock()
	cs.Storage[clientID] = &Client{
		RAM:       ram,
		CPU:       cpu,
		UpdatedAt: time.Now(),
	}
}

func (s *Storage) UpdateClient(id string, ram string, cpu string) {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	if client, exists := s.Storage[id]; exists {
		client.RAM = ram
		client.CPU = cpu
		client.UpdatedAt = time.Now()
	}
}

func (s *Storage) GetClient(id string) *Client {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	if client, exists := s.Storage[id]; exists {
		return client
	}

	return nil
}

func (cs *Storage) DeregisterInactiveClients(timeout time.Duration) {
	cs.Mu.Lock()
	defer cs.Mu.Unlock()
	for clientID, client := range cs.Storage {
		if time.Since(client.UpdatedAt) > timeout {
			delete(cs.Storage, clientID)
		}
	}
}
