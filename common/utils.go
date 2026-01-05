package common

import (
	"crypto/sha256"
	"fmt"
)

const (
	// ChunkSize is the size of each chunk in bytes (64MB)
	ChunkSize = 64 * 1024 * 1024

	// ReplicationFactor is the number of replicas for each chunk
	ReplicationFactor = 3

	// MasterAddress is the default master server address
	MasterAddress = "localhost:8000"
)

// GenerateChunkHandle generates a unique chunk handle based on filename and chunk index
func GenerateChunkHandle(filename string, chunkIndex int) string {
	data := fmt.Sprintf("%s-%d", filename, chunkIndex)
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash[:16])
}

// CalculateNumChunks calculates the number of chunks needed for a file
func CalculateNumChunks(filesize int64) int {
	numChunks := filesize / ChunkSize
	if filesize%ChunkSize != 0 {
		numChunks++
	}

	return int(numChunks)
}
