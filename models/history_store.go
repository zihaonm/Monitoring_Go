package models

import "sync"

// HistoryStore manages service history storage
type HistoryStore struct {
	histories map[string]*ServiceHistory
	mu        sync.RWMutex
	maxChecks int
}

// NewHistoryStore creates a new history store
func NewHistoryStore(maxChecks int) *HistoryStore {
	if maxChecks == 0 {
		maxChecks = 100
	}
	return &HistoryStore{
		histories: make(map[string]*ServiceHistory),
		maxChecks: maxChecks,
	}
}

// AddCheckResult adds a check result to service history
func (s *HistoryStore) AddCheckResult(serviceID string, record HealthCheckRecord) {
	s.mu.Lock()
	defer s.mu.Unlock()

	history, exists := s.histories[serviceID]
	if !exists {
		history = NewServiceHistory(serviceID, s.maxChecks)
		s.histories[serviceID] = history
	}

	history.AddCheck(record)
}

// GetHistory retrieves history for a service
func (s *HistoryStore) GetHistory(serviceID string) *ServiceHistory {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.histories[serviceID]
}

// GetStatistics retrieves statistics for a service
func (s *HistoryStore) GetStatistics(serviceID string) *ServiceStatistics {
	s.mu.RLock()
	defer s.mu.RUnlock()

	history, exists := s.histories[serviceID]
	if !exists {
		return &ServiceStatistics{
			ServiceID: serviceID,
		}
	}

	return history.GetStatistics()
}

// DeleteHistory removes history for a service
func (s *HistoryStore) DeleteHistory(serviceID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.histories, serviceID)
}

// GetAllAsMap returns all histories as a map (for persistence)
func (s *HistoryStore) GetAllAsMap() map[string]*ServiceHistory {
	s.mu.RLock()
	defer s.mu.RUnlock()

	historyCopy := make(map[string]*ServiceHistory)
	for k, v := range s.histories {
		historyCopy[k] = v
	}
	return historyCopy
}

// LoadFromMap loads histories from a map (from persistence)
func (s *HistoryStore) LoadFromMap(histories map[string]*ServiceHistory) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if histories != nil {
		s.histories = histories
	}
}
