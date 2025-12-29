package master

import (
	"slices"
	"sync"
	"time"
)

// FileMetadata represents metadata for a file
type FileMetadata struct {
	Filename   string
	Filesize   int64
	ChunkCount int
	Chunks     []string // chunk handles
	CreatedAt  time.Time
}

// ChunkMetadata represents metadata for a chunk
type ChunkMetadata struct {
	ChunkHandle string
	Locations   []string // chunk server addresses
	Version     int32
	Filename    string
	ChunkIndex  int32
}

// ChunkServerInfo represents a chunk server
type ChunkServerInfo struct {
	Address         string
	LatestHeartbeat time.Time
	Chunks          []string // chunk handles stored on this server
}

// Metadata manages all the metadata for the dfs
type Metadata struct {
	mu           sync.RWMutex
	files        map[string]*FileMetadata    // key: filename, value: file metadata
	chunks       map[string]*ChunkMetadata   // key: chunk handle, value: chunk metadata
	chunkServers map[string]*ChunkServerInfo // key: address, value: chunk server info
}

// NewMetadata creates a new metadata manager
func NewMetadata() *Metadata {
	return &Metadata{
		files:        make(map[string]*FileMetadata),
		chunks:       make(map[string]*ChunkMetadata),
		chunkServers: make(map[string]*ChunkServerInfo),
	}
}

// AddFile adds a new File to the metadata
func (m *Metadata) AddFile(filename string, filesize int64, chunkCount int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.files[filename] = &FileMetadata{
		Filename:   filename,
		Filesize:   filesize,
		ChunkCount: chunkCount,
		Chunks:     make([]string, 0, chunkCount),
		CreatedAt:  time.Now(),
	}
}

// AddChunkToFile adds a chunk handle to a file's chunk list
func (m *Metadata) AddChunkToFile(filename string, chunkHandle string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if file, exists := m.files[filename]; exists {
		file.Chunks = append(file.Chunks, chunkHandle)
	}
}

// AddChunk adds chunk metadata
func (m *Metadata) AddChunk(chunkHandle string, filename string, chunkIndex int32) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.chunks[chunkHandle] = &ChunkMetadata{
		ChunkHandle: chunkHandle,
		Locations:   make([]string, 0),
		Version:     1,
		Filename:    filename,
		ChunkIndex:  chunkIndex,
	}
}

// AddChunkLocation adds a chunk server location for a chunk
func (m *Metadata) AddChunkLocation(chunkHandle string, serverAddress string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if chunk, exists := m.chunks[chunkHandle]; exists {
		// if the location already exist then return to avoid duplicates
		if slices.Contains(chunk.Locations, serverAddress) {
			return
		}

		chunk.Locations = append(chunk.Locations, serverAddress)
	}
}

// GetFile fetches the file metadata
func (m *Metadata) GetFile(filename string) (*FileMetadata, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	file, exists := m.files[filename]
	return file, exists
}

// GetChunk fetches the chunk metadata
func (m *Metadata) GetChunk(chunkHandle string) (*ChunkMetadata, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	chunk, exists := m.chunks[chunkHandle]
	return chunk, exists
}

// ListFiles returns all the files
func (m *Metadata) ListFiles() []*FileMetadata {
	m.mu.RLock()
	defer m.mu.RUnlock()

	files := make([]*FileMetadata, 0, len(m.files))
	for _, file := range m.files {
		files = append(files, file)
	}

	return files
}

// RegisterChunkServer registers/update a chunk server
func (m *Metadata) RegisterChunkServer(address string, chunks []string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if server, exists := m.chunkServers[address]; exists {
		// update chunk server if server with given address exists
		server.LatestHeartbeat = time.Now()
		server.Chunks = chunks
	} else {
		// registers a new chunk server
		m.chunkServers[address] = &ChunkServerInfo{
			Address:         address,
			LatestHeartbeat: time.Now(),
			Chunks:          chunks,
		}
	}
}

// GetAvailableChunkServers returns the list of available chunk servers whose heartbeats had been updated recently within 30 secs
func (m *Metadata) GetAvailableChunkServers(replicationFactor int) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	servers := make([]string, 0, replicationFactor)
	now := time.Now()

	for address, server := range m.chunkServers {
		// only considers servers available if the heartbeat was updated within last 30 seconds
		if now.Sub(server.LatestHeartbeat) < 30*time.Second {
			servers = append(servers, address)
			if len(servers) >= replicationFactor {
				break
			}
		}
	}

	return servers
}

// GetAllChunkServers returns all registered chunk servers
func (m *Metadata) GetAllChunkServers() []string {
	m.mu.Lock()
	defer m.mu.Unlock()

	servers := make([]string, 0, len(m.chunkServers))
	for address := range m.chunkServers {
		servers = append(servers, address)
	}

	return servers
}
