package chunkserver

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	pb "github.com/harshvardha/distributed_file_system/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Server represents a chunk server
type Server struct {
	pb.UnimplementedChunkServerServer
	storage       *Storage
	address       string
	masterAddress string
}

// NewServer creates a new chunk server
func NewServer(address, storagePath, masterAddress string) (*Server, error) {
	storage, err := NewStorage(storagePath)
	if err != nil {
		return nil, err
	}

	return &Server{
		storage:       storage,
		address:       address,
		masterAddress: masterAddress,
	}, nil
}

// WriteChunk handles chunk write requests
func (s *Server) WriteChunk(ctx context.Context, req *pb.WriteChunkRequest) (*pb.WriteChunkResponse, error) {
	log.Printf("Writing chunk: %s (index: %d, size: %d bytes)", req.ChunkHandle, req.ChunkIndex, len(req.Data))

	if err := s.storage.WriteChunk(req.ChunkHandle, req.Data); err != nil {
		log.Printf("failed to write chunk %s to disk: %v", req.ChunkHandle, err)
		return &pb.WriteChunkResponse{Success: false}, err
	}

	// Reporting chunk storage to master
	go s.reportChunkToMaster(req.ChunkHandle)

	log.Printf("Successfully wrote chunk: %s to disk", req.ChunkHandle)
	return &pb.WriteChunkResponse{Success: true}, nil
}

// ReadChunk handles read chunk requests
func (s *Server) ReadChunk(ctx context.Context, req *pb.ReadChunkRequest) (*pb.ReadChunkResponse, error) {
	log.Printf("Reading chunk: %s from disk", req.ChunkHandle)

	data, err := s.storage.ReadChunk(req.ChunkHandle)
	if err != nil {
		log.Printf("failed to read chunk %s from disk: %v", req.ChunkHandle, err)
		return nil, err
	}

	log.Printf("Successfully read chunk %s with size %d from disk", req.ChunkHandle, len(data))
	return &pb.ReadChunkResponse{Data: data}, nil
}

// reportChunkToMaster reports chunk storage to master
func (s *Server) reportChunkToMaster(chunkHandle string) {
	conn, err := grpc.NewClient(s.masterAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("failed to connect to master: %v", err)
		return
	}

	defer conn.Close()

	client := pb.NewMasterClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = client.ReportChunk(ctx, &pb.ReportChunkRequest{
		ChunkHandle:        chunkHandle,
		ChunkServerAddress: s.address,
	})
	if err != nil {
		log.Printf("Chunk Server %s failed to report chunk storage to Master %s: %v", s.address, s.masterAddress, err)
	}
}

// startHeartbeat sends periodic heartbeats to master
func (s *Server) startHeartbeat() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		s.sendHeartbeat()
	}
}

// sendHeartbeat sends heartbeat to master
func (s *Server) sendHeartbeat() {
	conn, err := grpc.NewClient(s.masterAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("Failed to connect to master for sending heartbeat: %v", err)
		return
	}
	defer conn.Close()

	client := pb.NewMasterClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	chunks := s.storage.ListChunks()

	_, err = client.Heartbeat(ctx, &pb.HeartbeatRequest{
		ChunkServerAddress: s.address,
		ChunkHandles:       chunks,
	})

	if err != nil {
		log.Printf("Hearbeat failed: %v", err)
	} else {
		log.Printf("Heartbeat sent: %d chunks", len(chunks))
	}
}

// Start starts the chunk server
func (s *Server) Start() error {
	listen, err := net.Listen("tcp", s.address)
	if err != nil {
		return fmt.Errorf("chunk server %s failed to listen: %v", s.address, err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterChunkServerServer(grpcServer, s)

	// Starting heartbeat in background
	go s.startHeartbeat()

	log.Printf("chunk server starting on %s", s.address)
	log.Printf("Storage path: %s", s.storage.storagePath)
	log.Printf("Master address: %s", s.masterAddress)

	if err := grpcServer.Serve(listen); err != nil {
		return fmt.Errorf("failed to start chunk server %s: %v", s.address, err)
	}

	return nil
}
