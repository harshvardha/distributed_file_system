package chunkserver

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// Storage manages chunk storage on disk
type Storage struct {
	mu          sync.RWMutex
	storagePath string
	chunks      map[string]bool // key: chunk handle, value: exists(true/false)
}

// NewStorage creates a new storage manager
func NewStorage(storagePath string) (*Storage, error) {
	// Creating storage directory if it doesn't exist
	if err := os.MkdirAll(storagePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage dictionary: %v", err)
	}

	storage := &Storage{
		storagePath: storagePath,
		chunks:      make(map[string]bool),
	}

	// Loading existing chunks
	if err := storage.loadExistingChunks(); err != nil {
		return nil, fmt.Errorf("failed to load existing chunks: %v", err)
	}

	return storage, nil
}

// loadExistingChunks scans the storage directory for existing chunks
func (s *Storage) loadExistingChunks() error {
	files, err := os.ReadDir(s.storagePath)
	if err != nil {
		return err
	}

	for _, file := range files {
		if !file.IsDir() {
			chunkHandle := file.Name()
			s.chunks[chunkHandle] = true
		}
	}

	return nil
}

// WriteChunk writes chunk data to disk
func (s *Storage) WriteChunk(chunkHandle string, data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	chunkPath := filepath.Join(s.storagePath, chunkHandle)
	if err := os.WriteFile(chunkPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write chunk to disk: %v", err)
	}

	s.chunks[chunkHandle] = true
	return nil
}

// ReadChunk reads chunk data from disk
func (s *Storage) ReadChunk(chunkHandle string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.chunks[chunkHandle] {
		return nil, fmt.Errorf("chunk not found: %s", chunkHandle)
	}

	chunkPath := filepath.Join(s.storagePath, chunkHandle)
	data, err := os.ReadFile(chunkPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read chunk: %v", err)
	}

	return data, nil
}

// HasChunk checks if a chunk exists
func (s *Storage) HasChunk(chunkHandle string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.chunks[chunkHandle]
}

// ListChunks retuns all chunk handles
func (s *Storage) ListChunks() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	chunks := make([]string, 0, len(s.chunks))
	for chunkHandle := range s.chunks {
		chunks = append(chunks, chunkHandle)
	}

	return chunks
}

// DeleteChunk deletes a chunk from disk
func (s *Storage) DeleteChunk(chunkHandle string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	chunkPath := filepath.Join(s.storagePath, chunkHandle)

	if err := os.Remove(chunkPath); err != nil {
		return fmt.Errorf("failed to delete chunk: %v", err)
	}

	delete(s.chunks, chunkHandle)
	return nil
}
