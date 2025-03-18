package listener

import (
	"fmt"
	"sync"
)

// Registry manages data sources
type Registry struct {
	sources map[string]DataSource
	mu      sync.RWMutex
}

// NewRegistry creates a new registry
func NewRegistry() *Registry {
	return &Registry{
		sources: make(map[string]DataSource),
	}
}

// Register adds a data source to the registry
func (r *Registry) Register(id string, source DataSource) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.sources[id]; exists {
		return fmt.Errorf("data source with ID %s already exists", id)
	}

	r.sources[id] = source
	return nil
}

// Unregister removes a data source from the registry
func (r *Registry) Unregister(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.sources[id]; !exists {
		return fmt.Errorf("data source with ID %s does not exist", id)
	}

	delete(r.sources, id)
	return nil
}

// GetSource retrieves a data source by ID
func (r *Registry) GetSource(id string) (DataSource, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	source, exists := r.sources[id]
	if !exists {
		return nil, fmt.Errorf("data source with ID %s does not exist", id)
	}

	return source, nil
}

// GetAllSources returns all registered data sources
func (r *Registry) GetAllSources() map[string]DataSource {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Create a copy to avoid race conditions
	sources := make(map[string]DataSource, len(r.sources))
	for id, source := range r.sources {
		sources[id] = source
	}

	return sources
}
