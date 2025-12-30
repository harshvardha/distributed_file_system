package client

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/harshvardha/distributed_file_system/common"
	pb "github.com/harshvardha/distributed_file_system/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client represents a dfs client
type Client struct {
	masterAddress string
}

// NewClient creates a new DFS Client
func NewClient(masterAddress string) *Client {
	return &Client{
		masterAddress: masterAddress,
	}
}

// UploadFile uploads a file to the dfs
func (c *Client) UploadFile(localPath, remoteName string) error {
	log.Printf("Uploading file: %s as %s", localPath, remoteName)

	// Reading file
	data, err := os.ReadFile(localPath)
	if err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}

	filesize := int64(len(data))
	log.Printf("File size: %d bytes", filesize)

	// Creating a connection to master server
	conn, err := grpc.NewClient(c.masterAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect to master server: %v", err)
	}
	defer conn.Close()

	masterClient := pb.NewMasterClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Request chunk allocation
	response, err := masterClient.UploadFile(ctx, &pb.UploadFileRequest{
		Filename: remoteName,
		Filesize: filesize,
	})
	if err != nil {
		return fmt.Errorf("failed to request file upload: %v", err)
	}

	log.Printf("Recieved %d chunk locations", len(response.ChunkLocations))

	// Uploading chunks to chunk servers
	for _, chunkLoc := range response.ChunkLocations {
		if err := c.uploadChunk(data, chunkLoc); err != nil {
			return fmt.Errorf("failed to upload chunk %d: %v", chunkLoc.ChunkIndex, err)
		}
	}

	log.Printf("Successfully uploaded file: %s", remoteName)
	return nil
}

// uploadChunk uploads a single chunk to chunk servers
func (c *Client) uploadChunk(fileData []byte, chunkLoc *pb.ChunkLocation) error {
	// Calculating chunk data range
	chunkIndex := int(chunkLoc.ChunkIndex)
	start := chunkIndex * common.ChunkSize
	end := min(start+common.ChunkSize, len(fileData))

	chunkData := fileData[start:end]

	log.Printf("Uploading chunk %d (%s): %d bytes to %d servers", chunkIndex, chunkLoc.ChunkHandle, len(chunkData), len(chunkLoc.ChunkServerAddresses))

	// Upload to all replica servers
	for _, serverAddr := range chunkLoc.ChunkServerAddresses {
		if err := c.writeChunkToServer(serverAddr, chunkLoc.ChunkHandle, chunkData, chunkLoc.ChunkIndex); err != nil {
			log.Printf("Warning: failed to write chunk to %s: %v", serverAddr, err)
			// Continuing with other replicas
		} else {
			log.Printf("Successfully wrote chunk %d to %s", chunkIndex, serverAddr)
		}
	}

	return nil
}

// writeChunkToServer writes chunk data to a specific chunk server
func (c *Client) writeChunkToServer(serverAddr string, chunkHandle string, data []byte, chunkIndex int32) error {
	conn, err := grpc.NewClient(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect to chunk server %s: %v", serverAddr, err)
	}
	defer conn.Close()

	chunkClient := pb.NewChunkServerClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err = chunkClient.WriteChunk(ctx, &pb.WriteChunkRequest{
		ChunkHandle: chunkHandle,
		Data:        data,
		ChunkIndex:  chunkIndex,
	})

	return err
}

// DownloadFile downloads a file from the DFS
func (c *Client) DownloadFile(remoteName string, localPath string) error {
	log.Printf("Downloading file: %s to %s", remoteName, localPath)

	// Connecting to master server
	conn, err := grpc.NewClient(c.masterAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect to master server: %v", err)
	}
	defer conn.Close()

	masterClient := pb.NewMasterClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Requesting file metadata and chunk locations
	response, err := masterClient.DownloadFile(ctx, &pb.DownloadFileRequest{
		Filename: remoteName,
	})
	if err != nil {
		return fmt.Errorf("failed to request download: %v", err)
	}

	log.Printf("File size: %d bytes, %d chunks", response.Filesize, len(response.ChunkLocation))

	// Downloading chunks
	fileData := make([]byte, response.Filesize)
	for _, chunkLoc := range response.ChunkLocation {
		chunkData, err := c.downloadChunk(chunkLoc)
		if err != nil {
			return fmt.Errorf("failed to download chunk %d: %v", chunkLoc.ChunkIndex, err)
		}

		// Copying chunk data to file buffer
		chunkIndex := int(chunkLoc.ChunkIndex)
		start := chunkIndex * common.ChunkSize
		copy(fileData[start:], chunkData)
	}

	// Writing file to local disk
	if err := os.WriteFile(localPath, fileData, 0644); err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}

	log.Printf("Successfully downloaded file: %s", remoteName)
	return nil
}

// downloadChunk downloads a single chunk from the chunk servers
func (c *Client) downloadChunk(chunkLoc *pb.ChunkLocation) ([]byte, error) {
	log.Printf("Downloading chunk %d (%s) from %d servers", chunkLoc.ChunkIndex, chunkLoc.ChunkHandle, len(chunkLoc.ChunkServerAddresses))

	// Trying each server until on successfully downloads the chunk
	for _, serverAddr := range chunkLoc.ChunkServerAddresses {
		data, err := c.readChunkFromServer(serverAddr, chunkLoc.ChunkHandle)
		if err != nil {
			log.Printf("Warning: failed to read chunk from %s: %v", serverAddr, err)
			continue
		}

		log.Printf("Successfully read chunk %d from %s (%d bytes)", chunkLoc.ChunkIndex, serverAddr, len(data))
		return data, nil
	}

	return nil, fmt.Errorf("failed to download chunk from any server")
}

// readChunkFromServer reads chunk data from a specific chunk server
func (c *Client) readChunkFromServer(serverAddr, chunkHandle string) ([]byte, error) {
	conn, err := grpc.NewClient(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to chunk server: %v", err)
	}
	defer conn.Close()

	chunkClient := pb.NewChunkServerClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	response, err := chunkClient.ReadChunk(ctx, &pb.ReadChunkRequest{
		ChunkHandle: chunkHandle,
	})
	if err != nil {
		return nil, err
	}

	return response.Data, nil
}

// ListFiles lists all the files in the DFS
func (c *Client) ListFiles() ([]*pb.FileInfo, error) {
	log.Printf("Listing files...")

	// Connecting to master server
	conn, err := grpc.NewClient(c.masterAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to master server: %v", err)
	}
	defer conn.Close()

	masterClient := pb.NewMasterClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := masterClient.ListFiles(ctx, &pb.ListFilesRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %v", err)
	}

	return response.Files, nil
}
