package models

import (
	"errors"
	"sort"
	"sync"
)

var (
	ErrServiceNotFound = errors.New("service not found")
	ErrServiceExists   = errors.New("service already exists")
)

// ServiceStore manages the storage of monitored services
type ServiceStore struct {
	services    map[string]*MonitoredService
	mu          sync.RWMutex
	persistence *PersistenceManager
	onSave      func() // callback when data changes
}

// NewServiceStore creates a new service store
func NewServiceStore() *ServiceStore {
	return &ServiceStore{
		services: make(map[string]*MonitoredService),
	}
}

// SetPersistence sets the persistence manager and callback
func (s *ServiceStore) SetPersistence(pm *PersistenceManager, onSave func()) {
	s.persistence = pm
	s.onSave = onSave
}

// LoadFromMap loads services from a map (used during startup)
func (s *ServiceStore) LoadFromMap(services map[string]*MonitoredService) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.services = services
}

// GetAllAsMap returns all services as a map
func (s *ServiceStore) GetAllAsMap() map[string]*MonitoredService {
	s.mu.RLock()
	defer s.mu.RUnlock()

	servicesCopy := make(map[string]*MonitoredService)
	for k, v := range s.services {
		servicesCopy[k] = v
	}
	return servicesCopy
}

// triggerSave calls the onSave callback if set
func (s *ServiceStore) triggerSave() {
	if s.onSave != nil {
		go s.onSave()
	}
}

// Add adds a new service to the store
func (s *ServiceStore) Add(service *MonitoredService) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.services[service.ID]; exists {
		return ErrServiceExists
	}

	s.services[service.ID] = service
	s.triggerSave()
	return nil
}

// Get retrieves a service by ID
func (s *ServiceStore) Get(id string) (*MonitoredService, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	service, exists := s.services[id]
	if !exists {
		return nil, ErrServiceNotFound
	}

	return service, nil
}

// GetAll returns all monitored services sorted by creation date (newest first)
func (s *ServiceStore) GetAll() []*MonitoredService {
	s.mu.RLock()
	defer s.mu.RUnlock()

	services := make([]*MonitoredService, 0, len(s.services))
	for _, service := range s.services {
		services = append(services, service)
	}

	// Sort by CreatedAt in descending order (newest first)
	sort.Slice(services, func(i, j int) bool {
		return services[i].CreatedAt.After(services[j].CreatedAt)
	})

	return services
}

// Update updates an existing service
func (s *ServiceStore) Update(service *MonitoredService) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.services[service.ID]; !exists {
		return ErrServiceNotFound
	}

	s.services[service.ID] = service
	s.triggerSave()
	return nil
}

// Delete removes a service from the store
func (s *ServiceStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.services[id]; !exists {
		return ErrServiceNotFound
	}

	delete(s.services, id)
	s.triggerSave()
	return nil
}
